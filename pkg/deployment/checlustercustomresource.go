package deployment

import (
	che "github.com/eclipse/che-operator/pkg/apis/org/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCheClusterCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *che.CheCluster {
	return &che.CheCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CheCluster",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: che.CheClusterSpec{
			Server: che.CheClusterSpecServer{
				CheImageTag:          "",
				DevfileRegistryImage: "",
				PluginRegistryImage:  "",
				TlsSupport:           false,
				SelfSignedCert:       false,
			},
			Database: che.CheClusterSpecDB{
				ExternalDB:            false,
				ChePostgresDBHostname: "",
				ChePostgresPort:       "",
				ChePostgresUser:       "",
				ChePostgresPassword:   "",
				ChePostgresDb:         "",
			},
			Auth: che.CheClusterSpecAuth{
				OpenShiftOauth:   true,
				KeycloakImage:    "",
				ExternalKeycloak: false,
				KeycloakURL:      "",
				KeycloakRealm:    "",
				KeycloakClientId: "",
			},
			Storage: che.CheClusterSpecStorage{
				PvcStrategy:       "per-workspace",
				PvcClaimSize:      "1Gi",
				PreCreateSubPaths: true,
			},
		},
	}
}
