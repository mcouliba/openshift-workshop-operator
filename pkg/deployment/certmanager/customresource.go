package certmanager

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCertManagerCustomResource(name string, namespace string) *CertManager {

	return &CertManager{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertManager",
			APIVersion: "operator.cert-manager.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: CertManagerSpec{},
	}
}
