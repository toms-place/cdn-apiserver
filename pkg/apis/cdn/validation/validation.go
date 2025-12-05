/*
Copyright 2016 The Kubernetes Authors.

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

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.toms.place/apiserver/pkg/apis/cdn"
)

// ValidateFile validates a File.
func ValidateFile(f *cdn.File) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateFileSpec(&f.Spec, field.NewPath("spec"))...)

	return allErrs
}

// ValidateFileSpec validates a FileSpec.
func ValidateFileSpec(s *cdn.FileSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}
