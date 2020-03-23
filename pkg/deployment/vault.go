package deployment

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewVaultStateful(cr *openshiftv1alpha1.Workshop, name string, namespace string, labels map[string]string) *appsv1.StatefulSet {

	image := cr.Spec.Infrastructure.Vault.Image.Name + ":" + cr.Spec.Infrastructure.Vault.Image.Tag

	replicas := int32(1)
	terminationGracePeriodSeconds := int64(10)

	runAsNonRoot := true
	runAsGroup := int64(1000)
	runAsUser := int64(100)
	fsGroup := int64(1000)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         name + "-internal",
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &replicas,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: labels,
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					ServiceAccountName:            name,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
						RunAsGroup:   &runAsGroup,
						RunAsUser:    &runAsUser,
						FSGroup:      &fsGroup,
					},

					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name + "-config",
									},
								},
							},
						},
					},

					Containers: []corev1.Container{
						{
							Name: "vault",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"IPC_LOCK",
									},
								},
							},
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/sh",
								"-ec",
							},
							Args: []string{
								`sed -E "s/HOST_IP/${HOST_IP?}/g" /vault/config/extraconfig-from-values.hcl > /tmp/storageconfig.hcl;
sed -Ei "s/POD_IP/${POD_IP?}/g" /tmp/storageconfig.hcl;
/usr/local/bin/docker-entrypoint.sh vault server -config=/tmp/storageconfig.hcl`,
							},
							Env: []corev1.EnvVar{
								{
									Name: "HOST_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: "POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name:  "VAULT_ADDR",
									Value: "http://127.0.0.1:8200",
								},
								{
									Name:  "VAULT_API_ADDR",
									Value: "http-internal://$(POD_IP):8200",
								},
								{
									Name:  "SKIP_CHOWN",
									Value: "true",
								},
								{
									Name:  "SKIP_SETCAP",
									Value: "true",
								},
								{
									Name: "HOSTNAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/vault/data",
								},
								{
									Name:      "config",
									MountPath: "/vault/config",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8200,
								},
								{
									Name:          "internal",
									ContainerPort: 8201,
								},
								{
									Name:          "replication",
									ContainerPort: 8202,
								},
							},
							ReadinessProbe: &corev1.Probe{
								FailureThreshold:    2,
								InitialDelaySeconds: 5,
								PeriodSeconds:       3,
								SuccessThreshold:    1,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/sh",
											"-ec",
											"vault status -tls-skip-verify",
										},
									},
								},
							},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/sh",
											"-c",
											"sleep 5 && kill -SIGTERM $(pidof vault)",
										},
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		},
	}
}

func NewVaultAgentInjectorDeployment(cr *openshiftv1alpha1.Workshop, name string, namespace string, labels map[string]string) *appsv1.Deployment {
	image := cr.Spec.Infrastructure.Vault.AgentInjectorImage.Name + ":" + cr.Spec.Infrastructure.Vault.AgentInjectorImage.Tag
	vaultImage := cr.Spec.Infrastructure.Vault.Image.Name + ":" + cr.Spec.Infrastructure.Vault.Image.Tag

	replicas := int32(1)

	runAsNonRoot := true
	runAsGroup := int64(1000)
	runAsUser := int64(100)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: name,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
						RunAsGroup:   &runAsGroup,
						RunAsUser:    &runAsUser,
					},
					Containers: []corev1.Container{
						{
							Name:            "sidecar-injector",
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "AGENT_INJECT_LISTEN",
									Value: ":8080",
								},
								{
									Name:  "AGENT_INJECT_LOG_LEVEL",
									Value: "info",
								},
								{
									Name:  "AGENT_INJECT_VAULT_ADDR",
									Value: "http://vault." + namespace + ".svc:8200",
								},
								{
									Name:  "AGENT_INJECT_VAULT_AUTH_PATH",
									Value: "auth/kubernetes",
								},
								{
									Name:  "AGENT_INJECT_VAULT_IMAGE",
									Value: vaultImage,
								},
								{
									Name:  "AGENT_INJECT_TLS_AUTO",
									Value: name + "-cfg",
								},
								{
									Name:  "AGENT_INJECT_TLS_AUTO_HOSTS",
									Value: "vault-agent-injector,vault-agent-injector." + namespace + ",vault-agent-injector." + namespace + "svc",
								},
								{
									Name:  "AGENT_INJECT_LOG_FORMAT",
									Value: "standard",
								},
								{
									Name:  "AGENT_INJECT_REVOKE_ON_SHUTDOWN",
									Value: "false",
								},
							},
							Args: []string{
								"agent-inject",
								"2>&1",
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health/ready",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: int32(8080),
										},
										Scheme: corev1.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 1,
								FailureThreshold:    2,
								PeriodSeconds:       2,
								SuccessThreshold:    1,
								TimeoutSeconds:      5,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health/ready",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: int32(8080),
										},
										Scheme: corev1.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 1,
								FailureThreshold:    2,
								PeriodSeconds:       2,
								SuccessThreshold:    1,
								TimeoutSeconds:      5,
							},
						},
					},
				},
			},
		},
	}
}
