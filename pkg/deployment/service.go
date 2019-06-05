package deployment

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewService(cr *cloudnativev1alpha1.Workshop, name string, namespace string, portName []string, portNumber []int32) *corev1.Service {
	labels := GetLabels(cr, name)
	ports := []corev1.ServicePort{}
	for i := range portName {
		port := corev1.ServicePort{
			Name:     portName[i],
			Port:     portNumber[i],
			Protocol: "TCP",
		}
		ports = append(ports, port)
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: labels,
		},
	}
}
