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

package apiserver

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	"k8s.toms.place/apiserver/pkg/apis/cdn"
	cdninstall "k8s.toms.place/apiserver/pkg/apis/cdn/install"
	registry "k8s.toms.place/apiserver/pkg/registry"
	filestorage "k8s.toms.place/apiserver/pkg/registry/cdn/file"
)

var (
	// Scheme defines methods for serializing and deserializing API objects.
	Scheme = runtime.NewScheme()
	// Codecs provides methods for retrieving codecs and serializers for specific
	// versions and content types.
	Codecs           = serializer.NewCodecFactory(Scheme)
	CDNComponentName = "cdn"
)

func init() {
	cdninstall.Install(Scheme)

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

// ExtraConfig holds custom apiserver config
type ExtraConfig struct {
	// ExternalHost is the host used to construct URLs for file content endpoints.
	// If empty, the request's Host header will be used.
	ExternalHost string
}

// Config defines the config for the apiserver
type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// Server contains state for a Kubernetes cluster master/api server.
type Server struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

// CompletedConfig embeds a private pointer that cannot be instantiated outside of this package.
type CompletedConfig struct {
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	return CompletedConfig{&c}
}

// New returns a new instance of Server from the given config.
func (c completedConfig) New() (*Server, error) {
	genericServer, err := c.GenericConfig.New("apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &Server{
		GenericAPIServer: genericServer,
	}

	// Install CDN API group
	cdnAPIGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(cdn.GroupName, Scheme, metav1.ParameterCodec, Codecs)

	fileStorage := registry.RESTInPeace(filestorage.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter))
	cdnV1alpha1storage := map[string]rest.Storage{}
	cdnV1alpha1storage["files"] = fileStorage
	cdnV1alpha1storage["files/content"] = filestorage.NewContentREST(fileStorage, c.ExtraConfig.ExternalHost)
	cdnAPIGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = cdnV1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&cdnAPIGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
