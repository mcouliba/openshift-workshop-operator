package deployment

import (
	oauthv1 "github.com/openshift/api/oauth/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewOAuthClient(cr *openshiftv1alpha1.Workshop, name string, redirectURIs []string) *oauthv1.OAuthClient {
	return &oauthv1.OAuthClient{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OAuthClient",
			APIVersion: oauthv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		RedirectURIs: redirectURIs,
		Secret:       "openshift",
		GrantMethod:  oauthv1.GrantHandlerPrompt,
	}
}
