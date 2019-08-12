package customresource

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNexusCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *Nexus {
	return &Nexus{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Nexus",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: NexusSpec{
			NexusVolumeSize: "5Gi",
		},
	}
}
