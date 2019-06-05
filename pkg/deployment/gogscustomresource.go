package deployment

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	gogscustomresource "github.com/redhat/cloud-native-workshop-operator/pkg/customresource/gogs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewGogsCustomResource(cr *cloudnativev1alpha1.Workshop, name string, namespace string) *gogscustomresource.Gogs {
	labels := GetLabels(cr, name)
	return &gogscustomresource.Gogs{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gogs",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: gogscustomresource.GogsSpec{
			GogsVolumeSize:       "4Gi",
			GogsSsl:              false,
			PostgresqlVolumeSize: "4Gi",
		},
	}
}
