/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	"k8s.toms.place/apiserver/pkg/apis/cdn"
	cdnv1alpha1 "k8s.toms.place/apiserver/pkg/apis/cdn/v1alpha1"
	"k8s.toms.place/apiserver/pkg/registry"
)

// contentEntry holds the content and its last status
type contentEntry struct {
	data   []byte
	status metav1.Status
}

// contentStore is an in-memory store for file contents and their status
var contentStore = struct {
	sync.RWMutex
	entries map[string]*contentEntry
}{
	entries: make(map[string]*contentEntry),
}

// allowedMIMETypes defines the valid top-level MIME type categories
var allowedMIMETypes = map[string]bool{
	"application": true,
	"audio":       true,
	"font":        true,
	"image":       true,
	"message":     true,
	"model":       true,
	"multipart":   true,
	"text":        true,
	"video":       true,
}

// isValidMIMEType checks if the media type has a valid top-level type
func isValidMIMEType(mediaType string) bool {
	// Split the media type into type/subtype
	parts := strings.SplitN(mediaType, "/", 2)
	if len(parts) != 2 {
		return false
	}
	topLevel := parts[0]
	subType := parts[1]

	// Check if the top-level type is allowed
	if !allowedMIMETypes[topLevel] {
		return false
	}

	// Subtype must not be empty
	if subType == "" {
		return false
	}

	return true
}

// ContentREST implements rest.Connecter for streaming file content
type ContentREST struct {
	store        *registry.REST
	externalHost string
}

// NewContentREST creates a new ContentREST
// externalHost is optional - if empty, the request's Host header will be used
func NewContentREST(store *registry.REST, externalHost string) *ContentREST {
	return &ContentREST{
		store:        store,
		externalHost: externalHost,
	}
}

var _ rest.Connecter = &ContentREST{}
var _ rest.StorageMetadata = &ContentREST{}

// New returns an empty object that can be used with Create and Update
func (r *ContentREST) New() runtime.Object {
	return &cdn.FileContent{}
}

// Destroy cleans up resources on shutdown
func (r *ContentREST) Destroy() {}

// Connect returns an http.Handler that will stream the file content
func (r *ContentREST) Connect(ctx context.Context, name string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	opts, ok := options.(*cdn.FileContent)
	if !ok {
		return nil, fmt.Errorf("invalid options object: %#v", options)
	}

	return &contentHandler{
		ctx:          ctx,
		store:        r.store,
		name:         name,
		options:      opts,
		responder:    responder,
		externalHost: r.externalHost,
	}, nil
}

// NewConnectOptions returns an empty options object for the Connect method
func (r *ContentREST) NewConnectOptions() (runtime.Object, bool, string) {
	return &cdn.FileContent{}, false, ""
}

// ConnectMethods returns the list of HTTP methods handled by Connect
func (r *ContentREST) ConnectMethods() []string {
	return []string{"GET", "HEAD", "PUT"}
}

// ProducesMIMETypes returns a list of MIME types the verb can respond with
func (r *ContentREST) ProducesMIMETypes(verb string) []string {
	return []string{"application/octet-stream", "*/*"}
}

// ProducesObject returns the object the verb responds with
func (r *ContentREST) ProducesObject(verb string) interface{} {
	return nil
}

// contentHandler handles HTTP requests for file content streaming
type contentHandler struct {
	ctx          context.Context
	store        *registry.REST
	name         string
	options      *cdn.FileContent
	responder    rest.Responder
	externalHost string
}

// buildContentURL constructs the full URL for a file's content endpoint
// based on the configured external host (or request host as fallback) and the namespace/name from context
func (h *contentHandler) buildContentURL(req *http.Request) string {
	namespace := request.NamespaceValue(h.ctx)

	// Use configured external host, or fall back to request host
	host := h.externalHost
	if host == "" {
		host = req.Host
	}

	// Determine the scheme
	scheme := "https"
	if req.TLS == nil {
		scheme = "http"
	}

	// Build the URL: /apis/{group}/{version}/namespaces/{namespace}/files/{name}/content
	path := fmt.Sprintf("/apis/%s/%s/namespaces/%s/files/%s/content",
		cdnv1alpha1.GroupName,
		cdnv1alpha1.SchemeGroupVersion.Version,
		namespace,
		h.name,
	)

	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

// ServeHTTP handles GET, HEAD, and PUT requests for file content
func (h *contentHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.handleGet(w, req, false)
	case http.MethodHead:
		h.handleGet(w, req, true)
	case http.MethodPut:
		h.handlePut(w, req)
	default:
		http.Error(w, fmt.Sprintf("method %s not allowed", req.Method), http.StatusMethodNotAllowed)
	}
}

// handleGet streams the file content (or just headers if headOnly is true)
func (h *contentHandler) handleGet(w http.ResponseWriter, req *http.Request, headOnly bool) {
	// Get the File object from the store
	obj, err := h.store.Get(h.ctx, h.name, &metav1.GetOptions{})
	if err != nil {
		h.responder.Error(err)
		return
	}

	file, ok := obj.(*cdn.File)
	if !ok {
		h.responder.Error(fmt.Errorf("object is not a File"))
		return
	}

	// First check if we have content stored locally
	contentStore.RLock()
	entry, hasLocal := contentStore.entries[h.name]
	contentStore.RUnlock()

	if hasLocal {
		// Serve from local store
		contentType := file.Spec.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(entry.data)))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", h.name))
		w.WriteHeader(http.StatusOK)
		if !headOnly {
			w.Write(entry.data)
		}
		return
	}

	// No local content, return not found status
	h.responder.Error(apierrors.NewNotFound(cdn.Resource("file"), h.name))

}

// handlePut uploads content to the file
func (h *contentHandler) handlePut(w http.ResponseWriter, req *http.Request) {
	// Read the content first
	var buf bytes.Buffer
	_, err := io.Copy(&buf, req.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusInternalServerError)
		return
	}
	contentBytes := buf.Bytes()
	contentSize := int64(len(contentBytes))

	// Determine and validate content type from request header
	contentType := req.Header.Get("Content-Type")

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Parse and validate the MIME type
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid Content-Type: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the MIME type is a recognized type
	if !isValidMIMEType(mediaType) {
		http.Error(w, fmt.Sprintf("unsupported Content-Type: %s (must be a valid MIME type like text/*, application/*, image/*, etc.)", mediaType), http.StatusBadRequest)
		return
	}

	// Reconstruct a normalized content type (media type with charset if present)
	if charset, ok := params["charset"]; ok {
		contentType = fmt.Sprintf("%s; charset=%s", mediaType, charset)
	} else {
		contentType = mediaType
	}

	// Try to get the existing File
	obj, err := h.store.Get(h.ctx, h.name, &metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			h.responder.Error(err)
			return
		}

		// Build the URL for this file's content endpoint
		contentURL := h.buildContentURL(req)

		// File doesn't exist, create it
		newFile := &cdn.File{
			ObjectMeta: metav1.ObjectMeta{
				Name: h.name,
			},
			Spec: cdn.FileSpec{
				URL:         contentURL,
				Size:        contentSize,
				ContentType: contentType,
			},
			Status: cdn.FileStatus{
				Uploaded: true,
			},
		}

		_, err = h.store.Create(h.ctx, newFile, rest.ValidateAllObjectFunc, &metav1.CreateOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create file resource: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// File exists, update it with new size and content type
		file, ok := obj.(*cdn.File)
		if !ok {
			h.responder.Error(fmt.Errorf("object is not a File"))
			return
		}

		// Build the URL for this file's content endpoint
		contentURL := h.buildContentURL(req)

		// Update file spec with URL, size and content type
		file.Spec.URL = contentURL
		file.Spec.Size = contentSize
		file.Spec.ContentType = contentType
		file.Status.Uploaded = true
		file.Status.Error = ""

		_, _, err = h.store.Update(h.ctx, h.name, rest.DefaultUpdatedObjectInfo(file), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &metav1.UpdateOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to update file resource: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Build the status response
	status := metav1.Status{
		Status:  metav1.StatusSuccess,
		Message: fmt.Sprintf("content uploaded successfully for file %s (%d bytes, %s)", h.name, contentSize, contentType),
		Details: &metav1.StatusDetails{
			Name: h.name,
			Kind: "File",
		},
		Code: http.StatusCreated,
	}

	// Store the content and status
	contentStore.Lock()
	contentStore.entries[h.name] = &contentEntry{
		data:   contentBytes,
		status: status,
	}
	contentStore.Unlock()

	// Return success response using FileContent with Status
	response := &cdn.FileContent{
		Status: status,
	}
	h.responder.Object(http.StatusCreated, response)
}
