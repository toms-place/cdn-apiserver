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

package cdn

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FileList is a list of File objects.
type FileList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []File
}

// FileSpec is the specification of a File.
type FileSpec struct {
	// URL is the URL of the file.
	URL string
	// Size is the size of the file in bytes.
	Size int64
	// ContentType is the MIME type of the file.
	ContentType string
	// Add a resource location for the content
	ResourceLocation string
}

// FileStatus is the status of a File.
type FileStatus struct {
	// Uploaded is true if the file has been uploaded.
	Uploaded bool
	// Error is an error message if the file upload failed.
	Error string
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// File is an example type with a spec and a status.
type File struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   FileSpec
	Status FileStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FileContent is the content subresource for a File
type FileContent struct {
	metav1.TypeMeta
	Status metav1.Status
}
