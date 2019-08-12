package maistra

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceMeshControlPlaneCR(param NewServiceMeshControlPlaneCRParameters) *ServiceMeshControlPlane {
	return &ServiceMeshControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceMeshControlPlane",
			APIVersion: "istio.openshift.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      param.Name,
			Namespace: param.Namespace,
		},
		Spec: ServiceMeshControlPlaneSpec{
			Istio: IstioSpec{
				Global: GlobalSpec{
					Proxy: ProxySpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
				Gateways: GatewaysSpec{
					IstioEgressgateway: IstioEgressgatewaySpec{
						AutoscaleEnabled: false,
					},
					IstioIngressgateway: IstioIngressgatewaySpec{
						AutoscaleEnabled: false,
						IorEnabled:       true,
					},
				},
				Mixer: MixerSpec{
					Policy: PolicySpec{
						AutoscaleEnabled: false,
					},
					Telemetry: TelemetrySpec{
						AutoscaleEnabled: false,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("1G"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("4G"),
							},
						},
					},
				},
				Pilot: PilotSpec{
					AutoscaleEnabled: false,
					TraceSampling:    100.0,
				},
				Kiali: KialiSpec{
					Enabled: true,
					Hub:     "quay.io/kiali",
					Tag:     "v1.0.0",
				},
				Tracing: TracingSpec{
					Enabled: true,
					Jaeger: JaegerSpec{
						Tag:      "1.13.1",
						Template: "all-in-one",
					},
				},
			},
		},
	}
}
