codegen_cmd = 'UPDATE_API_KNOWN_VIOLATIONS=true API_KNOWN_VIOLATIONS_DIR=./hack ./hack/update-codegen.sh'
#compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o artifacts/simple-image/kube-sample-apiserver'

local_resource(
  'go-codegen',
  codegen_cmd,
  deps=['./pkg/apis'],
  ignore=['pkg/apis/**/zz_generated.*.go', './pkg/generated/**']
)

# local_resource(
#   'example-go-compile',
#   compile_cmd,
#   deps=['./pkg'],
#   resource_deps=['go-codegen']
# )

load('ext://nerdctl', 'nerdctl_build')
nerdctl_build(
    ref='k8s.toms.place/apiserver',
    context='.'
)
k8s_yaml(kustomize('artifacts/example'))
