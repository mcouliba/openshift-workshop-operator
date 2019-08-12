package maistra

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

///////////////////////////
/// FUNCTION PARAMETERS ///
///////////////////////////

type NewServiceMeshMemberRollCRParameters struct {
	Name      string
	Namespace string
	Members   []string
}

////////////
/// TYPE ///
////////////

type ServiceMeshMemberRoll struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceMeshMemberRollSpec `json:"spec"`
}

type ServiceMeshMemberRollSpec struct {
	Members []string `json:"members"`
}

type ServiceMeshMemberRollList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceMeshMemberRoll `json:"items"`
}
