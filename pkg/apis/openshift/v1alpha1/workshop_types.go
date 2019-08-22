package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkshopSpec defines the desired state of Workshop
// +k8s:openapi-gen=true
type WorkshopSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Users        int             `json:"users"`
	UserPassword string          `json:"userPassword"`
	Etherpad     EtherpadSpec    `json:"etherpad"`
	Gogs         GogsSpec        `json:"gogs"`
	Nexus        NexusSpec       `json:"nexus"`
	ServiceMesh  ServiceMeshSpec `json:"servicemesh"`
	Guide        GuideSpec       `json:"guide"`
	Che          CheSpec         `json:"che"`
	Squash       SquashSpec      `json:"squash"`
}

type EtherpadSpec struct {
	Enabled bool `json:"enabled"`
}

type GogsSpec struct {
	Enabled bool `json:"enabled"`
}

type NexusSpec struct {
	Enabled bool `json:"enabled"`
}

type ServiceMeshSpec struct {
	Enabled             bool   `json:"enabled"`
	JaegerOperatorImage string `json:"jaegerOperatorImage"`
	KialiOperatorImage  string `json:"kialiOperatorImage"`
	IstioOperatorImage  string `json:"istioOperatorImage"`
}

type GuideSpec struct {
	Enabled                     bool   `json:"enabled"`
	GitRepositoryLabPath        string `json:"gitRepositoryLabPath"`
	GitRepositoryLabReference   string `json:"gitRepositoryLabReference"`
	GitRepositoryGuidePath      string `json:"gitRepositoryGuidePath"`
	GitRepositoryGuideReference string `json:"gitRepositoryGuideReference"`
	GitRepositoryGuideContext   string `json:"gitRepositoryGuideContext"`
	GitRepositoryGuideFile      string `json:"gitRepositoryGuideFile"`
}

type CheSpec struct {
	Enabled bool `json:"enabled"`
}

type SquashSpec struct {
	Enabled bool `json:"enabled"`
}

// WorkshopStatus defines the observed state of Workshop
// +k8s:openapi-gen=true
type WorkshopStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	CheRunning string `json:"cheRunning"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Workshop is the Schema for the workshops API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Workshop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkshopSpec   `json:"spec,omitempty"`
	Status WorkshopStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkshopList contains a list of Workshop
type WorkshopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workshop `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workshop{}, &WorkshopList{})
}
