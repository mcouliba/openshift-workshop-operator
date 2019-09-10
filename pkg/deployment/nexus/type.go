package nexus

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type NexusSpec struct {
	NexusVolumeSize        string                       `json:"nexusVolumeSize"`
	NexusSSL               bool                         `json:"nexusSsl"`
	NexusImageTag          string                       `json:"nexusImageTag"`
	NexusCPURequest        int                          `json:"nexusCpuRequest"`
	NexusCPULimit          int                          `json:"nexusCpuLimit"`
	NexusMemoryRequest     string                       `json:"nexusMemoryRequest"`
	NexusMemoryLimit       string                       `json:"nexusMemoryLimit"`
	NexusReposMavenProxy   []NexusReposMavenProxySpec   `json:"nexus_repos_maven_proxy"`
	NexusReposMavenHosted  []NexusReposMavenHostedSpec  `json:"nexus_repos_maven_hosted"`
	NexusReposMavenGroup   []NexusReposMavenGroupSpec   `json:"nexus_repos_maven_group"`
	NexusReposDockerHosted []NexusReposDockerHostedSpec `json:"nexus_repos_docker_hosted"`
	NexusReposNpmProxy     []NexusReposNpmProxySpec     `json:"nexus_repos_npm_proxy"`
	NexusReposNpmGroup     []NexusReposNpmGroupSpec     `json:"nexus_repos_npm_group"`
}

type NexusReposMavenProxySpec struct {
	Name         string `json:"name"`
	RemoteURL    string `json:"remote_url"`
	LayoutPolicy string `json:"layout_policy"`
}

type NexusReposMavenHostedSpec struct {
	Name          string `json:"name"`
	VersionPolicy string `json:"version_policy"`
	WritePolicy   string `json:"write_policy"`
}

type NexusReposMavenGroupSpec struct {
	Name        string   `json:"name"`
	MemberRepos []string `json:"member_repos"`
}

type NexusReposDockerHostedSpec struct {
	Name      string `json:"name"`
	HttpPort  int    `json:"http_port"`
	V1Enabled bool   `json:"v1_enabled"`
}

type NexusReposNpmProxySpec struct {
	Name      string `json:"name"`
	RemoteURL string `json:"remote_url"`
}

type NexusReposNpmGroupSpec struct {
	Name        string   `json:"name"`
	MemberRepos []string `json:"member_repos"`
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
