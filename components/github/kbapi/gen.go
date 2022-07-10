package api

//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0 object crd:crdVersions=v1 output:crd:artifacts:config=config/ paths=./...

//+kubebuilder:object:generate=true
//+groupName=github.kweb.dev
//+versionName=v1alpha1
