#!/usr/bin/env bash

UPDATE_API_KNOWN_VIOLATIONS=true API_KNOWN_VIOLATIONS_DIR=./hack ./hack/update-codegen.sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o artifacts/simple-image/kube-sample-apiserver
nerdctl -n k8s.io build -t k8s.toms.place/apiserver:test ./artifacts/simple-image
kubectl apply -k ./artifacts/example
kubectl rollout restart deployments -n toms-place
echo "Build complete."



