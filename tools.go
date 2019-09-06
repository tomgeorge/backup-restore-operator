package tools

import (
	_ "k8s.io/code-generator"
	_ "k8s.io/code-generator/cmd/client-gen"
	_ "k8s.io/code-generator/cmd/conversion-gen"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "k8s.io/code-generator/cmd/informer-gen"
	_ "k8s.io/code-generator/cmd/lister-gen"
	_ "k8s.io/gengo/args"
	_ "k8s.io/kube-openapi"
	_ "k8s.io/kube-openapi/cmd/openapi-gen"
	_ "k8s.io/kue-openapi/cmd/openapi-gen"
	_ "sigs.k8s.io/controller-tools/pkg/crd/generator"
)
