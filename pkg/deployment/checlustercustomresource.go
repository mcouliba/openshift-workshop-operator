package deployment

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	checlustercustomresource "github.com/redhat/cloud-native-workshop-operator/pkg/customresource/checluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCheClusterCustomResource(cr *cloudnativev1alpha1.Workshop, name string, namespace string, cheImage string, cheImageTag string, tlsSupport bool, selfSignedCert bool, openShiftoAuth bool) *checlustercustomresource.CheCluster {
	labels := GetLabels(cr, name)
	return &checlustercustomresource.CheCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CheCluster",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: checlustercustomresource.CheClusterSpec{
			Server: checlustercustomresource.ServerSpec{
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
			Storage: checlustercustomresource.StorageSpec{
				PvcStrategy:  "per-workspace",
				PvcClaimSize: "1Gi",
			},
			Database: checlustercustomresource.DatabaseSpec{
				ExternalDb:          false,
				ChePostgresHostName: "",
				ChePostgresPort:     "",
				ChePostgresUser:     "",
				ChePostgresPassword: "",
				ChePostgresDb:       "",
			},
			Auth: checlustercustomresource.AuthSpec{
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
