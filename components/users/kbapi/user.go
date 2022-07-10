package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec,omitempty"`
}

type UserSpec struct {
	Email          string          `json:"email,omitempty"`
	LinkedAccounts []LinkedAccount `json:"linkedAccounts,omitempty"`
}

type LinkedAccount struct {
	ProviderID       string `json:"providerID,omitempty"`
	ProviderUserID   string `json:"providerUserID,omitempty"`
	ProviderUserName string `json:"providerUserName,omitempty"`
}
