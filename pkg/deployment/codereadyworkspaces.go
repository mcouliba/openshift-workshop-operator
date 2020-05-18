package deployment

import (
	che "github.com/eclipse/che-operator/pkg/apis/org/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type codereadyUser struct {
	Username    string       `json:"username"`
	Enabled     bool         `json:"enabled"`
	Email       string       `json:"email"`
	Credentials []Credential `json:"credentials"`
	ClientRoles ClientRoles  `json:"clientRoles"`
}

type Credential struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ClientRoles struct {
	RealmManagement []string `json:"realm-management"`
}

func NewCodeReadyUser(cr *openshiftv1alpha1.Workshop, username string, password string) *codereadyUser {
	return &codereadyUser{
		Username: username,
		Enabled:  true,
		Email:    username + "@none.com",
		Credentials: []Credential{
			{
				Type:  "password",
				Value: password,
			},
		},
		ClientRoles: ClientRoles{
			RealmManagement: []string{
				"user",
			},
		},
	}
}

func NewCodeReadyWorkspacesCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *che.CheCluster {

	pluginRegistryImage := cr.Spec.Infrastructure.CodeReadyWorkspace.PluginRegistryImage.Name +
		":" + cr.Spec.Infrastructure.CodeReadyWorkspace.PluginRegistryImage.Tag

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
					"CHE_INFRA_KUBERNETES_NAMESPACE_DEFAULT":                "<username>-workspace",
					"CHE_LIMITS_USER_WORKSPACES_RUN_COUNT":                  "2",
					"CHE_WORKSPACE_ACTIVITY__CHECK__SCHEDULER__PERIOD__S":   "-1",
					"CHE_WORKSPACE_ACTIVITY__CLEANUP__SCHEDULER__PERIOD__S": "-1",
				},
				DevfileRegistryImage: "",
				PluginRegistryImage:  pluginRegistryImage,
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
				OpenShiftoAuth:                cr.Spec.Infrastructure.CodeReadyWorkspace.OpenshiftOAuth,
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
