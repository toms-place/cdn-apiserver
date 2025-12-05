/*
Copyright 2024 The Kubernetes Authors.

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

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// IOStreams provides the standard streams for commands
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// UploadOptions holds the options for the upload command
type UploadOptions struct {
	IOStreams

	// File path to upload
	FilePath string
	// Name of the File resource in Kubernetes
	ResourceName string
	// Namespace
	Namespace string
	// Content type override
	ContentType string
	// Kubeconfig path
	KubeConfig string
	// Context to use
	Context string
	// Create the resource if it doesn't exist
	Create bool
}

// NewUploadOptions creates new UploadOptions with default values
func NewUploadOptions(streams IOStreams) *UploadOptions {
	return &UploadOptions{
		IOStreams: streams,
		Namespace: "default",
		Create:    true,
	}
}

// NewCmdUpload creates the upload command
func NewCmdUpload(streams IOStreams) *cobra.Command {
	o := NewUploadOptions(streams)

	cmd := &cobra.Command{
		Use:   "upload [file-path] [resource-name]",
		Short: "Upload a file to the CDN API",
		Long: `Upload a file to the files.cdn.k8s.toms.place API.

This command uploads the content of a local file to a File resource in the
CDN API server. If the File resource doesn't exist, it will be created.

The resource name is optional - if not provided, it will be derived from the
filename (e.g., "index.html" becomes "index.html" as the resource name).

The content type is automatically detected from the file extension.

Examples:
  # Upload a file (resource name derived from filename)
  kubectl cdn upload index.html

  # Upload a file with explicit resource name
  kubectl cdn upload index.html my-index

  # Upload with a specific content type override
  kubectl cdn upload data.json my-data --content-type application/json

  # Upload to a specific namespace
  kubectl cdn upload style.css -n my-namespace
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.FilePath = args[0]
			if len(args) > 1 {
				o.ResourceName = args[1]
			} else {
				// Derive resource name from filename
				o.ResourceName = filepath.Base(o.FilePath)
			}
			return o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "Namespace of the File resource")
	cmd.Flags().StringVar(&o.ContentType, "content-type", "", "Content-Type for the file (auto-detected if not specified)")
	cmd.Flags().StringVar(&o.KubeConfig, "kubeconfig", "", "Path to kubeconfig file")
	cmd.Flags().StringVar(&o.Context, "context", "", "Kubernetes context to use")
	cmd.Flags().BoolVar(&o.Create, "create", true, "Create the File resource if it doesn't exist")

	return cmd
}

// Run executes the upload command
func (o *UploadOptions) Run() error {
	// Read the file
	fileData, err := os.ReadFile(o.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", o.FilePath, err)
	}

	// Determine content type
	contentType := o.ContentType
	if contentType == "" {
		// Try to detect from file extension
		ext := filepath.Ext(o.FilePath)
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// Build kubernetes client config
	config, err := o.buildConfig()
	if err != nil {
		return fmt.Errorf("failed to build kubernetes config: %w", err)
	}

	// Create REST client for the CDN API
	cdnConfig := *config
	cdnConfig.APIPath = "/apis"
	cdnConfig.GroupVersion = &cdnGroupVersion
	cdnConfig.NegotiatedSerializer = cdnCodec

	client, err := rest.RESTClientFor(&cdnConfig)
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Upload the content using PUT to the content subresource
	url := fmt.Sprintf("/apis/cdn.k8s.toms.place/v1alpha1/namespaces/%s/files/%s/content",
		o.Namespace, o.ResourceName)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, "", bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	result := client.Put().
		AbsPath(url).
		SetHeader("Content-Type", contentType).
		Body(fileData).
		Do(context.Background())

	if err := result.Error(); err != nil {
		return fmt.Errorf("failed to upload content: %w", err)
	}

	// Parse response
	rawBody, err := result.Raw()
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Status metav1.Status `json:"status"`
	}
	if err := json.Unmarshal(rawBody, &response); err != nil {
		// Try to print as string if not JSON
		fmt.Fprintf(o.Out, "Upload completed: %s\n", string(rawBody))
		return nil
	}

	if response.Status.Status == metav1.StatusSuccess {
		fmt.Fprintf(o.Out, "✓ Successfully uploaded %s to %s/%s (%d bytes, %s)\n",
			o.FilePath, o.Namespace, o.ResourceName, len(fileData), contentType)
	} else {
		fmt.Fprintf(o.Out, "Upload response: %s\n", response.Status.Message)
	}

	return nil
}

// buildConfig creates the kubernetes client config
func (o *UploadOptions) buildConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if o.KubeConfig != "" {
		loadingRules.ExplicitPath = o.KubeConfig
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if o.Context != "" {
		configOverrides.CurrentContext = o.Context
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}

// GetOptions holds the options for the get command
type GetOptions struct {
	IOStreams

	// Name of the File resource
	ResourceName string
	// Namespace
	Namespace string
	// Output file path (optional)
	OutputPath string
	// Kubeconfig path
	KubeConfig string
	// Context to use
	Context string
}

// NewGetOptions creates new GetOptions with default values
func NewGetOptions(streams IOStreams) *GetOptions {
	return &GetOptions{
		IOStreams: streams,
		Namespace: "default",
	}
}

// NewCmdGet creates the get command
func NewCmdGet(streams IOStreams) *cobra.Command {
	o := NewGetOptions(streams)

	cmd := &cobra.Command{
		Use:   "get [resource-name]",
		Short: "Get file content from the CDN API",
		Long: `Get file content from the files.cdn.k8s.toms.place API.

This command retrieves the content of a File resource from the CDN API server.

Examples:
  # Get file content and print to stdout
  kubectl cdn get my-index

  # Save file content to a local file
  kubectl cdn get my-index -o index.html

  # Get from a specific namespace
  kubectl cdn get my-styles -n my-namespace
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.ResourceName = args[0]
			return o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "Namespace of the File resource")
	cmd.Flags().StringVarP(&o.OutputPath, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVar(&o.KubeConfig, "kubeconfig", "", "Path to kubeconfig file")
	cmd.Flags().StringVar(&o.Context, "context", "", "Kubernetes context to use")

	return cmd
}

// Run executes the get command
func (o *GetOptions) Run() error {
	// Build kubernetes client config
	config, err := o.buildConfig()
	if err != nil {
		return fmt.Errorf("failed to build kubernetes config: %w", err)
	}

	// Create REST client for the CDN API
	cdnConfig := *config
	cdnConfig.APIPath = "/apis"
	cdnConfig.GroupVersion = &cdnGroupVersion
	cdnConfig.NegotiatedSerializer = cdnCodec

	client, err := rest.RESTClientFor(&cdnConfig)
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Get the content using GET from the content subresource
	url := fmt.Sprintf("/apis/cdn.k8s.toms.place/v1alpha1/namespaces/%s/files/%s/content",
		o.Namespace, o.ResourceName)

	result := client.Get().
		AbsPath(url).
		Do(context.Background())

	if err := result.Error(); err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}

	rawBody, err := result.Raw()
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Output to file or stdout
	if o.OutputPath != "" {
		if err := os.WriteFile(o.OutputPath, rawBody, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", o.OutputPath, err)
		}
		fmt.Fprintf(o.ErrOut, "✓ Saved %d bytes to %s\n", len(rawBody), o.OutputPath)
	} else {
		io.Copy(o.Out, bytes.NewReader(rawBody))
	}

	return nil
}

// buildConfig creates the kubernetes client config
func (o *GetOptions) buildConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if o.KubeConfig != "" {
		loadingRules.ExplicitPath = o.KubeConfig
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if o.Context != "" {
		configOverrides.CurrentContext = o.Context
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}
