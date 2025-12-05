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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.toms.place/apiserver/pkg/apis/cdn"
	"k8s.toms.place/apiserver/pkg/apis/cdn/validation"
)

// NewStrategy creates and returns a fileStrategy instance
func NewStrategy(typer runtime.ObjectTyper) fileStrategy {
	return fileStrategy{typer, names.SimpleNameGenerator}
}

// GetAttrs returns labels.Set, fields.Set, and error in case the given runtime.Object is not a File
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	apiserver, ok := obj.(*cdn.File)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a File")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), SelectableFields(apiserver), nil
}

// MatchFile is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchFile(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// SelectableFields returns a field set that represents the object.
func SelectableFields(obj *cdn.File) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type fileStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (fileStrategy) NamespaceScoped() bool {
	return true
}

func (fileStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (fileStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (fileStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	file := obj.(*cdn.File)
	return validation.ValidateFile(file)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (fileStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string { return nil }

func (fileStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (fileStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (fileStrategy) Canonicalize(obj runtime.Object) {
}

func (fileStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	file := obj.(*cdn.File)
	return validation.ValidateFile(file)
}

// WarningsOnUpdate returns warnings for the given update.
func (fileStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
