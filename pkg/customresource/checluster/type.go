package customresource

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CheClusterSpec struct {
	Server   ServerSpec   `json:"server"`
	Storage  StorageSpec  `json:"storage"`
	Database DatabaseSpec `json:"database"`
	Auth     AuthSpec     `json:"auth"`
}

type ServerSpec struct {
	CheFlavor      string `json:"cheFlavor"`
	CheImage       string `json:"cheImage"`
	CheImageTag    string `json:"cheImageTag"`
	TlsSupport     bool   `json:"tlsSupport"`
	SelfSignedCert bool   `json:"selfSignedCert"`
	ProxyURL       string `json:"proxyURL,omitempty"`
	ProxyPort      string `json:"proxyPort,omitempty"`
	NonProxyHosts  string `json:"nonProxyHosts,omitempty"`
	ProxyUser      string `json:"proxyUser,omitempty"`
	ProxyPassword  string `json:"proxyPassword,omitempty"`
}

type StorageSpec struct {
	PvcStrategy  string `json:"pvcStrategy"`
	PvcClaimSize string `json:"pvcClaimSize"`
}

type DatabaseSpec struct {
	ExternalDb          bool   `json:"externalDb"`
	ChePostgresHostName string `json:"chePostgresHostName,omitempty"`
	ChePostgresPort     string `json:"chePostgresPort,omitempty"`
	ChePostgresUser     string `json:"chePostgresUser,omitempty"`
	ChePostgresPassword string `json:"chePostgresPassword,omitempty"`
	ChePostgresDb       string `json:"chePostgresDb,omitempty"`
}

type AuthSpec struct {
	OpenShiftoAuth                bool   `json:"openShiftoAuth"`
	ExternalIdentityProvider      bool   `json:"externalIdentityProvider"`
	IdentityProviderAdminUserName string `json:"identityProviderAdminUserName"`
	IdentityProviderPassword      string `json:"identityProviderPassword"`
	IdentityProviderURL           string `json:"identityProviderURL,omitempty"`
	IdentityProviderRealm         string `json:"identityProviderRealm,omitempty"`
	IdentityProviderClientId      string `json:"identityProviderClientId,omitempty"`
}

type CheCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CheClusterSpec `json:"spec"`
}

type CheClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CheCluster `json:"items"`
}
