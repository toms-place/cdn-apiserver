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
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ListOptions holds the options for the list command
type ListOptions struct {
	IOStreams

	// Namespace
	Namespace string
	// All namespaces
	AllNamespaces bool
	// Kubeconfig path
	KubeConfig string
	// Context to use
	Context string
}

// NewListOptions creates new ListOptions with default values
func NewListOptions(streams IOStreams) *ListOptions {
	return &ListOptions{
		IOStreams: streams,
		Namespace: "default",
	}
}

// NewCmdList creates the list command
func NewCmdList(streams IOStreams) *cobra.Command {
	o := NewListOptions(streams)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List files in the CDN API",
		Long: `List files from the files.cdn.k8s.toms.place API.

This command lists all File resources in the specified namespace.

Examples:
  # List files in the default namespace
  kubectl cdn list

  # List files in a specific namespace
  kubectl cdn list -n my-namespace

  # List files in all namespaces
  kubectl cdn list -A
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "Namespace of the File resources")
	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false, "List files in all namespaces")
	cmd.Flags().StringVar(&o.KubeConfig, "kubeconfig", "", "Path to kubeconfig file")
	cmd.Flags().StringVar(&o.Context, "context", "", "Kubernetes context to use")

	return cmd
}

// FileListResponse represents the API response for listing files
type FileListResponse struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []FileResponse `json:"items"`
}

// FileResponse represents a single file in the list
type FileResponse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FileSpecResponse   `json:"spec"`
	Status            FileStatusResponse `json:"status"`
}

// FileSpecResponse represents the spec of a file
type FileSpecResponse struct {
	URL         string `json:"url,omitempty"`
	Size        int64  `json:"size,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

// FileStatusResponse represents the status of a file
type FileStatusResponse struct {
	Uploaded bool   `json:"uploaded,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Run executes the list command
func (o *ListOptions) Run() error {
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

	// Build the URL
	var url string
	if o.AllNamespaces {
		url = "/apis/cdn.k8s.toms.place/v1alpha1/files"
	} else {
		url = fmt.Sprintf("/apis/cdn.k8s.toms.place/v1alpha1/namespaces/%s/files", o.Namespace)
	}

	result := client.Get().
		AbsPath(url).
		Do(context.Background())

	if err := result.Error(); err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	rawBody, err := result.Raw()
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var fileList FileListResponse
	if err := json.Unmarshal(rawBody, &fileList); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Print table output
	w := tabwriter.NewWriter(o.Out, 0, 0, 2, ' ', 0)
	if o.AllNamespaces {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tSIZE\tCONTENT-TYPE\tUPLOADED")
	} else {
		fmt.Fprintln(w, "NAME\tSIZE\tCONTENT-TYPE\tUPLOADED")
	}

	for _, file := range fileList.Items {
		uploaded := "false"
		if file.Status.Uploaded {
			uploaded = "true"
		}

		contentType := file.Spec.ContentType
		if contentType == "" {
			contentType = "-"
		}

		if o.AllNamespaces {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
				file.Namespace, file.Name, file.Spec.Size, contentType, uploaded)
		} else {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
				file.Name, file.Spec.Size, contentType, uploaded)
		}
	}
	w.Flush()

	return nil
}

// buildConfig creates the kubernetes client config
func (o *ListOptions) buildConfig() (*rest.Config, error) {
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
