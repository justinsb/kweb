package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true

type UserAuth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserAuthSpec `json:"spec,omitempty"`
}

type UserAuthSpec struct {
	UserID         string `json:"userID,omitempty"`
	ProviderID     string `json:"providerID,omitempty"`
	ProviderUserID string `json:"providerUserID,omitempty"`
}
