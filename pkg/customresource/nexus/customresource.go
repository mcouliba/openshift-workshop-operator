package customresource

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNexusCustomResource(cr *cloudnativev1alpha1.Workshop, name string, namespace string) *Nexus {
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
