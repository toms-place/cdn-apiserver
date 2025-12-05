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

package wardleinitializer_test

import (
	"context"
	"testing"
	"time"

	"k8s.io/apiserver/pkg/admission"
	initializer "k8s.toms.place/apiserver/pkg/admission/initializer"
	"k8s.toms.place/apiserver/pkg/generated/clientset/versioned/fake"
	informers "k8s.toms.place/apiserver/pkg/generated/informers/externalversions"
)

// TestWantsInternalWardleInformerFactory ensures that the informer factory is injected
// when the WantsInternalWardleInformerFactory interface is implemented by a plugin.
func TestWantsInternalWardleInformerFactory(t *testing.T) {
	cs := &fake.Clientset{}
	sf := informers.NewSharedInformerFactory(cs, time.Duration(1)*time.Second)
	target := initializer.New(sf)

	wantInformerFactory := &wantInternalInformerFactory{}
	target.Initialize(wantInformerFactory)
	if wantInformerFactory.sf != sf {
		t.Errorf("expected informer factory to be initialized")
	}
}

// wantInternalInformerFactory is a test stub that fulfills the WantsInternalInformerFactory interface
type wantInternalInformerFactory struct {
	sf informers.SharedInformerFactory
}

func (f *wantInternalInformerFactory) SetInternalInformerFactory(sf informers.SharedInformerFactory) {
	f.sf = sf
}
func (f *wantInternalInformerFactory) Admit(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	return nil
}
func (f *wantInternalInformerFactory) Handles(o admission.Operation) bool { return false }
func (f *wantInternalInformerFactory) ValidateInitialization() error      { return nil }

var _ admission.Interface = &wantInternalInformerFactory{}
var _ initializer.WantsInternalInformerFactory = &wantInternalInformerFactory{}
