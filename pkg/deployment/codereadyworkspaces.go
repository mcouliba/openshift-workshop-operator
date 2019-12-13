package deployment

import (
	che "github.com/eclipse/che-operator/pkg/apis/org/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCodeReadyWorkspacesCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *che.CheCluster {
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
				CheImageTag: "",
				CheFlavor:   "codeready",
				CustomCheProperties: map[string]string{
					"CHE_INFRA_KUBERNETES_PVC_WAIT__BOUND": "false",
				},
				DevfileRegistryImage: "",
				PluginRegistryImage:  "quay.io/mcouliba/che-plugin-registry:7.3.x",
				TlsSupport:           true,
				SelfSignedCert:       false,
			},
			Database: che.CheClusterSpecDB{
				ExternalDb:          false,
				ChePostgresHostName: "",
				ChePostgresPort:     "",
				ChePostgresUser:     "",
				ChePostgresPassword: "",
				ChePostgresDb:       "",
			},
			Auth: che.CheClusterSpecAuth{
				OpenShiftoAuth:                true,
				IdentityProviderImage:         "",
				ExternalIdentityProvider:      false,
				IdentityProviderURL:           "",
				IdentityProviderRealm:         "",
				IdentityProviderClientId:      "",
				IdentityProviderAdminUserName: "admin",
				IdentityProviderPassword:      "admin",
			},
			Storage: che.CheClusterSpecStorage{
				PvcStrategy:       "per-workspace",
				PvcClaimSize:      "1Gi",
				PreCreateSubPaths: true,
			},
		},
	}
}
