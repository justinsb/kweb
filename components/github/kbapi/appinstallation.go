package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true

type AppInstallation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AppInstallationSpec `json:"spec,omitempty"`
}

type AppInstallationSpec struct {
	ID      int64          `json:"id,omitempty"`
	UserID  string         `json:"userID,omitempty"`
	Account *GithubAccount `json:"account,omitempty"`
}

type GithubAccount struct {
	ID    int64  `json:"id,omitempty"`
	Login string `json:"login,omitempty"`
}
