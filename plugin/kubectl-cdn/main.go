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

package main

import (
	"os"

	"github.com/spf13/cobra"

	"k8s.toms.place/apiserver/plugin/kubectl-cdn/pkg/cmd"
)

func main() {
	streams := cmd.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	rootCmd := &cobra.Command{
		Use:   "kubectl-cdn",
		Short: "kubectl plugin for managing CDN files",
		Long:  "A kubectl plugin to upload and manage file content in the files.cdn.k8s.toms.place API",
	}

	rootCmd.AddCommand(cmd.NewCmdUpload(streams))
	rootCmd.AddCommand(cmd.NewCmdGet(streams))
	rootCmd.AddCommand(cmd.NewCmdList(streams))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
