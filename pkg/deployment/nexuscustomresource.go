package deployment

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	nexuscustomresource "github.com/redhat/cloud-native-workshop-operator/pkg/customresource/nexus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNexusCustomResource(cr *cloudnativev1alpha1.Workshop, name string, namespace string) *nexuscustomresource.Nexus {
	labels := GetLabels(cr, name)
	return &nexuscustomresource.Nexus{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Nexus",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: nexuscustomresource.NexusSpec{
			NexusVolumeSize: "5Gi",
		},
	}
}
