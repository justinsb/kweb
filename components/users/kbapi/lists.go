package api

// TODO: Auto-generate?

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//+kubebuilder:object:root=true

// UserAuthList contains a list of UserAuth
type UserAuthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserAuth `json:"items"`
}

//+kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "kweb.dev", Version: "v1alpha1"}

	// We removed SchemeBuilder to keep our dependencies small

	KindUser = KindInfo{
		Resource: GroupVersion.WithResource("users"),
		objects:  []runtime.Object{&User{}, &UserList{}},
	}
	KindUserAuth = KindInfo{
		Resource: GroupVersion.WithResource("userauths"),
		objects:  []runtime.Object{&UserAuth{}, &UserAuthList{}},
	}

	AllKinds = []KindInfo{KindUser, KindUserAuth}
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
