package customresource

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type NexusSpec struct {
	NexusVolumeSize string `json:"nexusVolumeSize"`
}

type Nexus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NexusSpec `json:"spec"`
}

type NexusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Nexus `json:"items"`
}
