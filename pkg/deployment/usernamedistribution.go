package deployment

import (
	"fmt"
	"strconv"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewUsernameDistributionDeployment(cr *openshiftv1alpha1.Workshop, name string, namespace string, labels map[string]string,
	redisServiceName string, users int, appsHostnameSuffix string) *appsv1.Deployment {

	image := "quay.io/mcouliba/username-distribution:1.3"

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: name,
							Env: []corev1.EnvVar{
								{
									Name:  "LAB_REDIS_HOST",
									Value: redisServiceName,
								},
								{
									Name:  "LAB_REDIS_PASS",
									Value: redisServiceName,
								},
								{
									Name:  "LAB_TITLE",
									Value: "OpenShift Workshops",
								},
								{
									Name:  "LAB_DURATION_HOURS",
									Value: "1week",
								},
								{
									Name:  "LAB_USER_COUNT",
									Value: strconv.Itoa(users),
								},
								{
									Name:  "LAB_USER_ACCESS_TOKEN",
									Value: cr.Name,
								},
								{
									Name:  "LAB_USER_PASS",
									Value: cr.Spec.User.Password,
								},
								{
									Name:  "LAB_USER_PREFIX",
									Value: "user",
								},
								{
									Name:  "LAB_USER_PAD_ZERO",
									Value: "false",
								},
								{
									Name:  "LAB_ADMIN_PASS",
									Value: "r3dh4t1!",
								},
								{
									Name:  "LAB_MODULE_URLS",
									Value: fmt.Sprintf("http://%%USERNAME%%-bookbag-workshop-guides.%s/workshop", appsHostnameSuffix) + ";" + cr.Name,
								},
							},
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      "TCP",
								},
							},
						},
					},
				},
			},
		},
	}
}
