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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var cdnGroupVersion = schema.GroupVersion{Group: "cdn.k8s.toms.place", Version: "v1alpha1"}

var cdnScheme = runtime.NewScheme()
var cdnCodec = serializer.NewCodecFactory(cdnScheme)

func init() {
	// Register the types we need for the API
	metav1.AddToGroupVersion(cdnScheme, cdnGroupVersion)
}
