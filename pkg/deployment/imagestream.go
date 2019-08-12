package deployment

import (
	imagev1 "github.com/openshift/api/image/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewImageStream(cr *openshiftv1alpha1.Workshop, name string, namespace string, imageName string, imageVersion string) *imagev1.ImageStream {
	labels := GetLabels(cr, name)
	return &imagev1.ImageStream{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: imagev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name: imageVersion,
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: imageName,
					},
				},
			},
		},
	}
}
