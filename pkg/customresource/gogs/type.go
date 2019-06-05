package customresource

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type GogsSpec struct {
	GogsVolumeSize       string `json:"gogsVolumeSize"`
	GogsSsl              bool   `json:"gogsSsl"`
	GogsServiceName      string `json:"gogsServiceName,omitempty"`
	PostgresqlVolumeSize string `json:"postgresqlVolumeSize"`
}

type Gogs struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GogsSpec `json:"spec"`
}

type GogsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Gogs `json:"items"`
}
