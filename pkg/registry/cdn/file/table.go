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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	"k8s.toms.place/apiserver/pkg/apis/cdn"
)

type fileTableConvertor struct{}

var _ rest.TableConvertor = fileTableConvertor{}

func (fileTableConvertor) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	var table metav1.Table

	table.ColumnDefinitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Size", Type: "integer", Description: "Size of the file in bytes"},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		// Wide columns (Priority: 1 means only shown with -o wide)
		{Name: "Content-Type", Type: "string", Priority: 1, Description: "MIME type of the file"},
		{Name: "Uploaded", Type: "boolean", Priority: 1, Description: "Whether the file has been uploaded"},
	}

	switch obj := object.(type) {
	case *cdn.FileList:
		table.ResourceVersion = obj.ResourceVersion
		table.Continue = obj.Continue
		for i := range obj.Items {
			table.Rows = append(table.Rows, fileToRow(&obj.Items[i]))
		}
	case *cdn.File:
		table.ResourceVersion = obj.ResourceVersion
		table.Rows = append(table.Rows, fileToRow(obj))
	}

	return &table, nil
}

func fileToRow(file *cdn.File) metav1.TableRow {
	return metav1.TableRow{
		Object: runtime.RawExtension{Object: file},
		Cells: []interface{}{
			file.Name,
			file.Spec.Size,
			translateTimestampSince(file.CreationTimestamp),
			// Wide columns (kubectl filters based on Priority)
			file.Spec.ContentType,
			file.Status.Uploaded,
		},
	}
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return humanDuration(time.Since(timestamp.Time))
}

// humanDuration returns a human-readable approximation of a duration
// (eg. "About a minute", "4 hours ago", etc.).
func humanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < -1 {
		return "<invalid>"
	} else if seconds < 0 {
		return "0s"
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}
