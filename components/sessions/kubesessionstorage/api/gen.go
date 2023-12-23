package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0 object crd:crdVersions=v1 output:crd:artifacts:config=config/ paths=./...

//+kubebuilder:object:generate=true
//+groupName=kweb.dev
//+versionName=v1alpha1

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "kweb.dev", Version: "v1alpha1"}

	// We removed SchemeBuilder to keep our dependencies small

	KindUser = KindInfo{
		Resource: GroupVersion.WithResource("sessions"),
		objects:  []runtime.Object{&Session{}, &SessionList{}},
	}

	AllKinds = KindInfoList{KindUser}
)

//+kubebuilder:object:generate=false

// KindInfo holds type meta-information
type KindInfo struct {
	Resource schema.GroupVersionResource
	objects  []runtime.Object
}

// GroupResource returns the GroupResource for the kind
func (k *KindInfo) GroupResource() schema.GroupResource {
	return k.Resource.GroupResource()
}

//+kubebuilder:object:generate=false

type KindInfoList []KindInfo

func (l KindInfoList) AddToScheme(scheme *runtime.Scheme) error {
	for _, kind := range l {
		scheme.AddKnownTypes(GroupVersion, kind.objects...)
	}
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
