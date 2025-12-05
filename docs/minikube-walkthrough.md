# CDN API Server Walkthrough

This document will take you through setting up and trying the CDN API server on a local Kubernetes cluster (minikube, kind, or Rancher Desktop).

## Prerequisites

- Go 1.25+ installed and setup. More information can be found at [Go installation](https://go.dev/doc/install)
- A local Kubernetes cluster:
  - [minikube](https://minikube.sigs.k8s.io/docs/start/)
  - [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
  - [Rancher Desktop](https://rancherdesktop.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed
- Docker or containerd for building images

## Clone the Repository

```bash
git clone https://github.com/kubernetes/sample-apiserver.git
cd sample-apiserver
```

## Option 1: Quick Start with Tilt (Recommended)

If you have [Tilt](https://tilt.dev/) installed, the fastest way to get started is:

```bash
tilt up
```

This will automatically:

- Generate the API code
- Build the container image
- Deploy all resources to your cluster
- Watch for changes and rebuild automatically

Skip to [Test the CDN API](#test-the-cdn-api) section.

## Option 2: Manual Setup

### Generate API Code

First, run the code generation scripts to generate clients, informers, and listers:

```bash
./hack/update-codegen.sh
```

### Build the Binary

From the root of the repo, build the API server binary:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o artifacts/simple-image/kube-sample-apiserver .
```

If everything went well, you should have a binary in `artifacts/simple-image/kube-sample-apiserver`.

### Build the Container Image

Build and push the Docker image:

```bash
docker build -t <YOUR_REGISTRY>/cdn-apiserver:latest ./artifacts/simple-image
docker push <YOUR_REGISTRY>/cdn-apiserver:latest
```

For local development with minikube, you can load the image directly:

```bash
# For minikube
eval $(minikube docker-env)
docker build -t k8s.toms.place/apiserver:test ./artifacts/simple-image

# For kind
kind load docker-image k8s.toms.place/apiserver:test
```

### Update the Deployment (if using custom registry)

If you pushed to a custom registry, update the image in `artifacts/example/apiserver/deployment.yaml`:

```yaml
containers:
  - name: api-server
    image: <YOUR_REGISTRY>/cdn-apiserver:latest
    imagePullPolicy: Always # Change from Never if pulling from registry
```

### Deploy to Kubernetes

Deploy all resources using kustomize:

```bash
kubectl apply -k artifacts/example
```

This will create:

- Namespace `toms-place`
- ServiceAccount and RBAC rules
- Deployment with the API server and etcd sidecar
- Service for the API server
- APIService registration for `cdn.k8s.toms.place/v1alpha1`

Verify the deployment:

```bash
# Check the namespace was created
kubectl get ns toms-place

# Check the pods are running
kubectl get pods -n toms-place

# Check the APIService is available
kubectl get apiservice v1alpha1.cdn.k8s.toms.place
```

Wait until the API server pod is running and the APIService shows `Available: True`.

## Test the CDN API

### Create a File Resource

The CDN API server registers the `File` resource type under `cdn.k8s.toms.place/v1alpha1`.

Create your first File resource:

```bash
kubectl apply -f artifacts/test-resources/file.yaml
```

Or create one directly:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: cdn.k8s.toms.place/v1alpha1
kind: File
metadata:
  name: my-first-file.txt
  namespace: default
spec:
  url: https://example.com/my-first-file.txt
  size: 1024
  contentType: text/plain
EOF
```

### List Files

```bash
# List files in default namespace
kubectl get files

# List files in all namespaces
kubectl get files -A

# Get detailed output
kubectl get files -o wide
```

### Get File Details

```bash
kubectl get file my-first-file -o yaml
```

Expected output:

```yaml
apiVersion: cdn.k8s.toms.place/v1alpha1
kind: File
metadata:
  name: my-first-file.txt
  namespace: default
spec:
  url: https://example.com/my-first-file.txt
  size: 1024
  contentType: text/plain
status:
  uploaded: false
  error: ""
```

### Delete a File

```bash
kubectl delete file my-first-file
```

## Using the kubectl Plugin

For a better CLI experience, install the kubectl-cdn plugin:

```bash
cd plugin/kubectl-cdn
go build -o kubectl-cdn .
cp kubectl-cdn /usr/local/bin/
```

Then use it:

```bash
# List files
kubectl cdn list

# Upload a local file
kubectl cdn upload /path/to/local/file.txt my-uploaded-file

# Get file details
kubectl cdn get my-uploaded-file
```

See [plugin/kubectl-cdn/README.md](../plugin/kubectl-cdn/README.md) for full usage.

## Cleanup

To remove all deployed resources:

```bash
kubectl delete -k artifacts/example
```

## Troubleshooting

### APIService shows Available: False

Check the API server logs:

```bash
kubectl logs -n toms-place -l apiserver=true
```

Common issues:

- etcd not ready - wait for the etcd container to start
- Image pull errors - verify the image exists and is accessible

### Cannot create File resources

Verify the APIService is registered and available:

```bash
kubectl get apiservice v1alpha1.cdn.k8s.toms.place -o yaml
```

Check that the service is reachable:

```bash
kubectl get svc -n toms-place
kubectl get endpoints -n toms-place
```

### Permission denied errors

Ensure RBAC is properly configured:

```bash
kubectl get clusterrolebinding | grep apiserver
kubectl get rolebinding -n kube-system | grep auth
```
