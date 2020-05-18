package certmanager

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CertManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CertManagerSpec `json:"spec,omitempty"`
}

type CertManagerSpec struct{}

type CertManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CertManager `json:"items"`
}
