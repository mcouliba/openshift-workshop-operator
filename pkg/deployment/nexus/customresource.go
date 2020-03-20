package nexus

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *Nexus {
	return &Nexus{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Nexus",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: NexusSpec{
			NexusVolumeSize:    "5Gi",
			NexusSSL:           true,
			NexusImageTag:      "3.18.1-01-ubi-3",
			NexusCPURequest:    1,
			NexusCPULimit:      2,
			NexusMemoryRequest: "2Gi",
			NexusMemoryLimit:   "2Gi",
			NexusReposMavenProxy: []NexusReposMavenProxySpec{
				{
					Name:         "maven-central",
					RemoteURL:    "https://repo1.maven.org/maven2/",
					LayoutPolicy: "permissive",
				},
				{
					Name:         "redhat-ga",
					RemoteURL:    "https://maven.repository.redhat.com/ga/",
					LayoutPolicy: "permissive",
				},
				{
					Name:         "jboss",
					RemoteURL:    "https://repository.jboss.org/nexus/content/groups/public",
					LayoutPolicy: "permissive",
				},
			},
			NexusReposMavenHosted: []NexusReposMavenHostedSpec{
				{
					Name:          "releases",
					VersionPolicy: "release",
					WritePolicy:   "allow_once",
				},
			},
			NexusReposMavenGroup: []NexusReposMavenGroupSpec{
				{
					Name:        "maven-all-public",
					MemberRepos: []string{"maven-central", "redhat-ga", "jboss"},
				},
			},
			NexusReposDockerHosted: []NexusReposDockerHostedSpec{
				{
					Name:      "docker",
					HttpPort:  5000,
					V1Enabled: true,
				},
			},
			NexusReposNpmProxy: []NexusReposNpmProxySpec{
				{
					Name:      "npm",
					RemoteURL: "https://registry.npmjs.org",
				},
			},
			NexusReposNpmGroup: []NexusReposNpmGroupSpec{
				{
					Name:        "npm-all",
					MemberRepos: []string{"npm"},
				},
			},
		},
	}
}
