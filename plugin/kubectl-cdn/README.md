# kubectl-cdn

A kubectl plugin to upload and manage file content in the `files.cdn.k8s.toms.place` Kubernetes API.

## Installation

### Build from source

```bash
cd plugin/kubectl-cdn
go build -o kubectl-cdn .
```

### Install as kubectl plugin

Copy the binary to a directory in your PATH with the `kubectl-` prefix:

```bash
# macOS/Linux
cp kubectl-cdn /usr/local/bin/kubectl-cdn

# Or add to your PATH
export PATH="$PATH:$(pwd)"
```

Once installed, you can use the plugin via kubectl:

```bash
kubectl cdn upload myfile.txt my-file
kubectl cdn get my-file
```

## Usage

### List files

List all File resources:

```bash
# List files in the default namespace
kubectl cdn list

# List files in a specific namespace
kubectl cdn list -n my-namespace

# List files in all namespaces
kubectl cdn list -A
```

### Upload a file

Upload a local file to the CDN API:

```bash
# Upload a file (auto-detects content type)
kubectl cdn upload index.html my-index

# Upload with explicit content type
kubectl cdn upload data.json my-data --content-type application/json

# Upload to a specific namespace
kubectl cdn upload style.css my-styles -n my-namespace
```

### Get file content

Retrieve file content from the CDN API:

```bash
# Print content to stdout
kubectl cdn get my-index

# Save to a local file
kubectl cdn get my-index -o index.html

# Get from a specific namespace
kubectl cdn get my-styles -n my-namespace
```

## Flags

### Common flags

| Flag           | Short | Description                                         |
| -------------- | ----- | --------------------------------------------------- |
| `--namespace`  | `-n`  | Namespace of the File resource (default: "default") |
| `--kubeconfig` |       | Path to kubeconfig file                             |
| `--context`    |       | Kubernetes context to use                           |

### Upload-specific flags

| Flag             | Description                                                  |
| ---------------- | ------------------------------------------------------------ |
| `--content-type` | Content-Type for the file (auto-detected if not specified)   |
| `--create`       | Create the File resource if it doesn't exist (default: true) |

### Get-specific flags

| Flag       | Short | Description                        |
| ---------- | ----- | ---------------------------------- |
| `--output` | `-o`  | Output file path (default: stdout) |

### List-specific flags

| Flag               | Short | Description                  |
| ------------------ | ----- | ---------------------------- |
| `--all-namespaces` | `-A`  | List files in all namespaces |

## How it works

This plugin interacts with the `files.cdn.k8s.toms.place/v1alpha1` API, specifically the `/content` subresource of `File` resources.

- **Upload**: Sends a PUT request to `/apis/cdn.k8s.toms.place/v1alpha1/namespaces/{namespace}/files/{name}/content`
- **Get**: Sends a GET request to the same endpoint

The API server stores the file content and updates the File resource metadata (size, content type, upload status).

## Examples

### Upload an HTML file

```bash
echo "<h1>Hello World</h1>" > index.html
kubectl cdn upload index.html my-homepage
```

### Upload a JSON configuration

```bash
kubectl cdn upload config.json my-config --content-type application/json -n prod
```

### Download and view content

```bash
# View content directly
kubectl cdn get my-homepage

# Save to file
kubectl cdn get my-homepage -o downloaded.html
```

## Development

### Building

```bash
go build -o kubectl-cdn .
```

### Testing

```bash
go test ./...
```
