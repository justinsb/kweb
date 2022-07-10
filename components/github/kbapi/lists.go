package api

// TODO: Auto-generate?

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//+kubebuilder:object:root=true

// AppInstallationList contains a list of Installation
type AppInstallationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppInstallation `json:"items"`
}

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "github.kweb.dev", Version: "v1alpha1"}

	// We removed SchemeBuilder to keep our dependencies small

	KindInstallation = KindInfo{
		Resource: GroupVersion.WithResource("installations"),
		objects:  []runtime.Object{&AppInstallation{}, &AppInstallationList{}},
	}

	AllKinds = []KindInfo{KindInstallation}
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

func AddToScheme(scheme *runtime.Scheme) error {
	for _, kind := range AllKinds {
		scheme.AddKnownTypes(GroupVersion, kind.objects...)
	}
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
