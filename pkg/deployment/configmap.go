package deployment

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewConfigMap(cr *cloudnativev1alpha1.Workshop, name string, namespaces string, data map[string]string) *corev1.ConfigMap {
	labels := GetLabels(cr, name)
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespaces,
			Labels:    labels,
		},
		Data: data,
	}
}
