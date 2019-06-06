package customresource

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCheClusterCustomResource(cr *cloudnativev1alpha1.Workshop, name string, namespace string, cheImage string, cheImageTag string, tlsSupport bool, selfSignedCert bool, openShiftoAuth bool) *CheCluster {
	return &CheCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CheCluster",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: CheClusterSpec{
			Server: ServerSpec{
				CheFlavor:      "codeready",
				CheImage:       cheImage,
				CheImageTag:    cheImageTag,
				TlsSupport:     tlsSupport,
				SelfSignedCert: selfSignedCert,
				ProxyURL:       "",
				ProxyPort:      "",
				NonProxyHosts:  "",
				ProxyUser:      "",
				ProxyPassword:  "",
			},
			Storage: StorageSpec{
				PvcStrategy:  "per-workspace",
				PvcClaimSize: "1Gi",
			},
			Database: DatabaseSpec{
				ExternalDb:          false,
				ChePostgresHostName: "",
				ChePostgresPort:     "",
				ChePostgresUser:     "",
				ChePostgresPassword: "",
				ChePostgresDb:       "",
			},
			Auth: AuthSpec{
				OpenShiftoAuth:                openShiftoAuth,
				ExternalIdentityProvider:      false,
				IdentityProviderAdminUserName: "admin",
				IdentityProviderPassword:      "admin",
				IdentityProviderURL:           "",
				IdentityProviderRealm:         "",
				IdentityProviderClientId:      "",
			},
		},
	}
}
