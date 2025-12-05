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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// FileSpec is the specification of a File.
type FileSpec struct {
	// URL is the URL of the file.
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// Size is the size of the file in bytes.
	Size int64 `json:"size,omitempty" protobuf:"varint,2,opt,name=size"`
	// ContentType is the MIME type of the file.
	ContentType string `json:"contentType,omitempty" protobuf:"bytes,3,opt,name=contentType"`
	// Add a resource location for the content
	ResourceLocation string `json:"resourceLocation,omitempty" protobuf:"bytes,4,opt,name=resourceLocation"`
}

// FileStatus is the status of a File.
type FileStatus struct {
	// Uploaded is true if the file has been uploaded.
	Uploaded bool `json:"uploaded,omitempty" protobuf:"varint,1,opt,name=uploaded"`
	// Error is an error message if the file upload failed.
	Error string `json:"error,omitempty" protobuf:"bytes,2,opt,name=error"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease-lifecycle-gen:introduced=1.0
// +k8s:prerelease-lifecycle-gen:removed=1.10

type File struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              FileSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            FileStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease-lifecycle-gen:introduced=1.0
// +k8s:prerelease-lifecycle-gen:removed=1.10

// FileList is a list of File objects.
type FileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []File `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease-lifecycle-gen:introduced=1.0
// +k8s:prerelease-lifecycle-gen:removed=1.10

// FileContent is the content subresource for a File
type FileContent struct {
	metav1.TypeMeta `json:",inline"`
	Status          metav1.Status `json:"status,omitempty" protobuf:"bytes,1,opt,name=status"`
}
