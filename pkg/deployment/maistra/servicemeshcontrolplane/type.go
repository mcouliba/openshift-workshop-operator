package maistra

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

///////////////////////////
/// FUNCTION PARAMETERS ///
///////////////////////////

type NewServiceMeshControlPlaneCRParameters struct {
	Name      string
	Namespace string
}

////////////
/// TYPE ///
////////////

type ServiceMeshControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceMeshControlPlaneSpec `json:"spec"`
}

type ServiceMeshControlPlaneSpec struct {
	Istio IstioSpec `json:"istio"`
}

type IstioSpec struct {
	Global   GlobalSpec   `json:"global"`
	Gateways GatewaysSpec `json:"gateways"`
	Mixer    MixerSpec    `json:"mixer"`
	Pilot    PilotSpec    `json:"pilot"`
	Kiali    KialiSpec    `json:"kiali"`
	Grafana  GrafanaSpec  `json:"grafana"`
	Tracing  TracingSpec  `json:"tracing"`
}

type GlobalSpec struct {
	Proxy ProxySpec `json:"proxy"`
}

type ProxySpec struct {
	Resources corev1.ResourceRequirements `json:"resources"`
}

type GatewaysSpec struct {
	IstioEgressgateway  IstioEgressgatewaySpec  `json:"istio-egressgateway"`
	IstioIngressgateway IstioIngressgatewaySpec `json:"istio-ingressgateway"`
}

type IstioEgressgatewaySpec struct {
	AutoscaleEnabled bool `json:"autoscaleEnabled"`
}

type IstioIngressgatewaySpec struct {
	AutoscaleEnabled bool `json:"autoscaleEnabled"`
	IorEnabled       bool `json:"ior_enabled"`
}

type MixerSpec struct {
	Policy    PolicySpec    `json:"policy"`
	Telemetry TelemetrySpec `json:"telemetry"`
}

type PolicySpec struct {
	AutoscaleEnabled bool `json:"autoscaleEnabled"`
}

type TelemetrySpec struct {
	AutoscaleEnabled bool                        `json:"autoscaleEnabled"`
	Resources        corev1.ResourceRequirements `json:"resources"`
}

type PilotSpec struct {
	AutoscaleEnabled bool    `json:"autoscaleEnabled"`
	TraceSampling    float32 `json:"traceSampling"`
}

type GrafanaSpec struct {
	Enabled bool `json:"enabled"`
}

type KialiSpec struct {
	Enabled bool `json:"enabled"`
}

type TracingSpec struct {
	Enabled bool       `json:"enabled"`
	Jaeger  JaegerSpec `json:"jaeger"`
}

type JaegerSpec struct {
	Template string `json:"template"`
}

type ServiceMeshControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceMeshControlPlane `json:"items"`
}
