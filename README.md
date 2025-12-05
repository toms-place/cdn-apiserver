# Sample API Server - CDN File Management

A Kubernetes sample API server that demonstrates how to build a custom aggregated API server with file/CDN management capabilities. Based on the [kubernetes/sample-apiserver](https://github.com/kubernetes/sample-apiserver) project.

## Overview

This project showcases how to extend Kubernetes with a custom API server that provides:

- **Custom Resource Definition (CRD)**: A `File` resource type under the `cdn.k8s.toms.place` API group
- **File Management API**: Upload, store, and serve files through Kubernetes-native APIs
- **Content Subresource**: Direct file content access via a `/content` subresource endpoint

## Architecture

```text
┌─────────────────────────────────────────────────────────────────┐
│                        Components                                │
├─────────────────┬─────────────────────┬─────────────────────────┤
│   API Server    │   kubectl Plugin    │      Web UI (Next.js)   │
│     (Go)        │      (Go)           │       (TypeScript)      │
├─────────────────┴─────────────────────┴─────────────────────────┤
│                    Kubernetes Cluster                            │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              cdn.k8s.toms.place/v1alpha1                │    │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                  │    │
│  │  │  File   │  │  File   │  │  File   │  ...             │    │
│  │  └─────────┘  └─────────┘  └─────────┘                  │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. API Server (`/pkg`)

A custom Kubernetes aggregated API server written in Go that:

- Registers the `cdn.k8s.toms.place` API group
- Provides CRUD operations for `File` resources
- Implements a `/content` subresource for file content retrieval
- Uses etcd for storage (via Kubernetes API machinery)

**File Resource Schema:**

```yaml
apiVersion: cdn.k8s.toms.place/v1alpha1
kind: File
metadata:
  name: example-file
spec:
  url: "https://example.com/file.txt"
  size: 1024
  contentType: "text/plain"
  resourceLocation: "/path/to/content"
status:
  uploaded: true
  error: ""
```

### 2. kubectl Plugin (`/plugin/kubectl-cdn`)

A kubectl plugin for managing files from the command line:

```bash
# Upload a file
kubectl cdn upload myfile.txt

# List files
kubectl cdn list
kubectl cdn list -n my-namespace
kubectl cdn list -A

# Get file details
kubectl cdn get myfile.txt
```

### 3. Web UI (`/app`)

A Next.js web application for viewing and managing File resources through a browser interface.

## Prerequisites

- Go 1.25+
- Kubernetes cluster (minikube, kind, Rancher Desktop, etc.)
- Docker/containerd for building images
- [Tilt](https://tilt.dev/) for local development (optional)

## Quick Start

### Local Development with Tilt

1. Start your Kubernetes cluster (e.g., Rancher Desktop, minikube)

2. Run Tilt:

   ```bash
   tilt up
   ```

   This will:

   - Run code generation for API types
   - Build the container image
   - Deploy to your cluster using kustomize

### Manual Build

1. **Generate code:**

   ```bash
   ./hack/update-codegen.sh
   ```

2. **Build the binary:**

   ```bash
   CGO_ENABLED=0 GOOS=linux go build -o artifacts/simple-image/kube-sample-apiserver
   ```

3. **Build container image:**

   ```bash
   docker build -t your-registry/apiserver:latest ./artifacts/simple-image
   docker push your-registry/apiserver:latest
   ```

4. **Deploy to cluster:**

   ```bash
   kubectl apply -k artifacts/example
   ```

### Install kubectl Plugin

```bash
cd plugin/kubectl-cdn
go build -o kubectl-cdn .
cp kubectl-cdn /usr/local/bin/
```

## Project Structure

```text
├── main.go                 # API server entrypoint
├── pkg/
│   ├── apis/cdn/          # API type definitions
│   │   ├── types.go       # File, FileSpec, FileStatus types
│   │   ├── v1alpha1/      # Versioned API
│   │   └── validation/    # Validation logic
│   ├── apiserver/         # Server configuration
│   ├── registry/          # Storage implementations
│   └── generated/         # Generated clients, informers, listers
├── plugin/kubectl-cdn/    # kubectl plugin
├── app/                   # Next.js web UI
├── artifacts/             # Kubernetes manifests & Dockerfile
├── hack/                  # Build and codegen scripts
└── vendor/                # Go dependencies
```

## API Reference

### File Resource

| Field                   | Type   | Description                    |
| ----------------------- | ------ | ------------------------------ |
| `spec.url`              | string | URL of the file                |
| `spec.size`             | int64  | File size in bytes             |
| `spec.contentType`      | string | MIME type of the file          |
| `spec.resourceLocation` | string | Internal resource location     |
| `status.uploaded`       | bool   | Whether file has been uploaded |
| `status.error`          | string | Error message if upload failed |

### Endpoints

- `GET /apis/cdn.k8s.toms.place/v1alpha1/namespaces/{ns}/files` - List files
- `GET /apis/cdn.k8s.toms.place/v1alpha1/namespaces/{ns}/files/{name}` - Get file
- `POST /apis/cdn.k8s.toms.place/v1alpha1/namespaces/{ns}/files` - Create file
- `PUT /apis/cdn.k8s.toms.place/v1alpha1/namespaces/{ns}/files/{name}` - Update file
- `DELETE /apis/cdn.k8s.toms.place/v1alpha1/namespaces/{ns}/files/{name}` - Delete file
- `GET /apis/cdn.k8s.toms.place/v1alpha1/namespaces/{ns}/files/{name}/content` - Get file content

## Documentation

- [Minikube Walkthrough](docs/minikube-walkthrough.md) - Step-by-step guide for local setup

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## Acknowledgments

Based on the [Kubernetes Sample API Server](https://github.com/kubernetes/sample-apiserver) project by the Kubernetes Authors.
