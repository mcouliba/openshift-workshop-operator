package deployment

import (
	"fmt"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewBookbagDeployment(cr *openshiftv1alpha1.Workshop, name string, namespace string, labels map[string]string,
	userID string, appsHostnameSuffix string, openshiftConsoleURL string) *appsv1.Deployment {

	user := fmt.Sprintf("user%s", userID)
	image := cr.Spec.Infrastructure.Bookbag.Image.Name + ":" + cr.Spec.Infrastructure.Bookbag.Image.Tag
	consoleImage := "quay.io/openshift/origin-console:4.2"

	vars := `{
	"OPENSHIFT_CONSOLE_URL": "` + openshiftConsoleURL + `",
	"APPS_HOSTNAME_SUFFIX": "` + appsHostnameSuffix + `",
	"USER_ID": "` + userID + `",
	"OPENSHIFT_PASSWORD": "` + cr.Spec.User.Password + `",
	"CHE_URL": "http://codeready-workspaces.` + appsHostnameSuffix + `",
	"GIT_URL": "https://gogs-gogs-server-workshop-infra.` + appsHostnameSuffix + `",
	"JAEGER_URL": "https://jaeger-istio-system.` + appsHostnameSuffix + `",
	"KIALI_URL": "https://kiali-istio-system.` + appsHostnameSuffix + `",
	"KIBANA_URL": "https://kibana-openshift-logging.` + appsHostnameSuffix + `",
	"GITOPS_URL": "https://argocd-server-argocd.` + appsHostnameSuffix + `",
	"WORKSHOP_GIT_REPO": "` + cr.Spec.Source.GitURL + `",
	"WORKSHOP_GIT_REF": "` + cr.Spec.Source.GitBranch + `"
}`

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
					ServiceAccountName: name,
					Volumes: []corev1.Volume{
						{
							Name: "envvars",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name + "-env",
									},
								},
							},
						},
						{
							Name: "workshopvars",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name + "-vars",
									},
								},
							},
						},
						{
							Name: "shared",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name: "setup-console",
							Command: []string{
								"/opt/workshop/bin/setup-console.sh",
							},
							Env: []corev1.EnvVar{
								{
									Name: "CLUSTER_SUBDOMAIN",
								},
								{
									Name: "OPENSHIFT_PROJECT",
								},
								{
									Name: "OPENSHIFT_USERNAME",
								},
								{
									Name: "OPENSHIFT_PASSWORD",
								},
								{
									Name: "OPENSHIFT_TOKEN",
								},
								{
									Name: "OC_VERSION",
								},
								{
									Name: "ODO_VERSION",
								},
								{
									Name: "KUBECTL_VERSION",
								},
							},
							Image: image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shared",
									MountPath: "/var/run/workshop",
								},
								{
									Name:      "workshopvars",
									MountPath: "/var/run/workshop-vars",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "terminal",
							Env: []corev1.EnvVar{
								{
									Name:  "APPLICATION_NAME",
									Value: name,
								},
								{
									Name:  "AUTH_USERNAME",
									Value: user,
								},
								{
									Name:  "AUTH_PASSWORD",
									Value: cr.Spec.User.Password,
								},
								{
									Name:  "OAUTH_SERVICE_ACCOUNT",
									Value: "bookbag",
								},
								{
									Name: "DOWNLOAD_URL",
								},
								{
									Name: "WORKSHOP_FILE",
								},
								{
									Name:  "CONSOLE_URL",
									Value: "http://0.0.0.0:10083",
								},
								{
									Name: "OC_VERSION",
								},
								{
									Name: "ODO_VERSION",
								},
								{
									Name: "KUBECTL_VERSION",
								},
								{
									Name:  "WORKSHOP_VARS",
									Value: vars,
								},
							},
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 10080,
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shared",
									MountPath: "/var/run/workshop",
								},
								{
									Name:      "workshopvars",
									MountPath: "/var/run/workshop-vars",
								},
							},
						},
						{
							Name: "console",
							Command: []string{
								"/var/run/workshop/start-console.sh",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "BRIDGE_K8S_MODE",
									Value: "in-cluster",
								},
								{
									Name:  "BRIDGE_LISTEN",
									Value: "http://0.0.0.0:10083",
								},
								{
									Name:  "BRIDGE_BASE_PATH",
									Value: "/console/",
								},
								{
									Name:  "BRIDGE_PUBLIC_DIR",
									Value: "/opt/bridge/static",
								},
								{
									Name:  "BRIDGE_USER_AUTH",
									Value: "disabled",
								},
								{
									Name:  "BRIDGE_BRANDING",
									Value: "openshift",
								},
							},
							Image: consoleImage,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shared",
									MountPath: "/var/run/workshop",
								},
							},
						},
					},
				},
			},
		},
	}
}
