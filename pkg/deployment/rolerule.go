package deployment

import (
	rbac "k8s.io/api/rbac/v1"
)

func NexusRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"namespaces",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
				"daemonsets",
				"replicasets",
				"statefulsets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"servicemonitors",
			},
			Verbs: []string{
				"get",
				"create",
			},
		},
		{
			APIGroups: []string{
				"gpte.opentlc.com",
			},
			Resources: []string{
				"nexus",
				"nexus/status",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			ResourceNames: []string{
				"nexus-operator",
			},
			Resources: []string{
				"deployments/finalizers",
			},
			Verbs: []string{
				"update",
			},
		},
	}
}

func GogsRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"namespaces",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
				"daemonsets",
				"replicasets",
				"statefulsets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"servicemonitors",
			},
			Verbs: []string{
				"get",
				"create",
			},
		},
		{
			APIGroups: []string{
				"gpte.opentlc.com",
			},
			Resources: []string{
				"gogs",
				"gogs/status",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			ResourceNames: []string{
				"gogs-operator",
			},
			Resources: []string{
				"deployments/finalizers",
			},
			Verbs: []string{
				"update",
			},
		},
	}
}

func WorkspacesRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"extensions/v1beta1",
			},
			Resources: []string{
				"ingresses",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"roles",
				"rolebindings",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"clusterroles",
				"clusterrolebindings",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"serviceaccounts",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
				"pods/exec",
				"pods/log",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"namespaces",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"servicemonitors",
			},
			Verbs: []string{
				"get",
				"create",
			},
		},
		{
			APIGroups: []string{
				"org.eclipse.che",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"oauth.openshift.io",
			},
			Resources: []string{
				"oauthclients",
			},
			Verbs: []string{
				"get",
				"create",
				"delete",
				"update",
				"list",
				"watch",
			},
		},
	}
}

func JaegerRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
				"serviceaccounts",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
				"daemonsets",
				"replicasets",
				"statefulsets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"servicemonitors",
			},
			Verbs: []string{
				"get",
				"create",
			},
		},
		{
			APIGroups: []string{
				"io.jaegertracing",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"jaegertracing.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"extensions",
			},
			Resources: []string{
				"ingresses",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"batch",
			},
			Resources: []string{
				"jobs",
				"cronjobs",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"logging.openshift.io",
			},
			Resources: []string{
				"elasticsearches",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
	}
}

func KialiRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"serviceaccounts",	
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"namespaces",
			},
			Verbs: []string{
				"get",
				"list",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
				"replicasets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"servicemonitors",
			},
			Verbs: []string{
				"create",
				"get",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			ResourceNames: []string{
				"kiali-operator",
			},
			Resources: []string{
				"deployments/finalizers",
			},
			Verbs: []string{
				"update",
			},
		},
		{
			APIGroups: []string{
				"kiali.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"clusterrolebindings",
				"clusterroles",
				"rolebindings",
				"roles",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apiextensions.k8s.io",
			},
			Resources: []string{
				"customresourcedefinitions",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"extensions",
			},
			Resources: []string{
				"ingresses",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"oauth.openshift.io",
			},
			Resources: []string{
				"oauthclients",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"monitoring.kiali.io",
			},
			Resources: []string{
				"monitoringdashboards",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"configmaps",
				"endpoints",
				"namespaces",
				"nodes",
				"pods",
				"pods/log",
				"replicationcontrollers",
				"services",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"extensions",
				"apps",
			},
			Resources: []string{
				"deployments",
				"replicasets",
				"statefulsets",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"autoscaling",
			},
			Resources: []string{
				"horizontalpodautoscalers",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"batch",
			},
			Resources: []string{
				"cronjobs",
				"jobs",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"config.istio.io",
			},
			Resources: []string{
				"adapters",
				"apikeys",
				"bypasses",
				"authorizations",
				"checknothings",
				"circonuses",
				"cloudwatches",
				"deniers",
				"dogstatsds",
				"edges",
				"fluentds",
				"handlers",
				"instances",
				"kubernetesenvs",
				"kuberneteses",
				"listcheckers",
				"listentries",
				"logentries",
				"memquotas",
				"metrics",
				"noops",
				"opas",
				"prometheuses",
				"quotas",
				"quotaspecbindings",
				"quotaspecs",
				"rbacs",
				"redisquotas",
				"reportnothings",
				"rules",
				"signalfxs",
				"solarwindses",
				"stackdrivers",
				"statsds",
				"stdios",
				"templates",
				"tracespans",
				"zipkins",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"networking.istio.io",
			},
			Resources: []string{
				"destinationrules",
				"gateways",
				"bypasses",
				"serviceentries",
				"sidecars",
				"virtualservices",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"authentication.istio.io",
			},
			Resources: []string{
				"meshpolicies",
				"policies",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"rbac.istio.io",
			},
			Resources: []string{
				"clusterrbacconfigs",
				"rbacconfigs",
				"servicerolebindings",
				"serviceroles",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps.openshift.io",
			},
			Resources: []string{
				"deploymentconfigs",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"project.openshift.io",
			},
			Resources: []string{
				"projects",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"monitoring.kiali.io",
			},
			Resources: []string{
				"monitoringdashboards",
			},
			Verbs: []string{
				"get",
				"list",
			},
		},
	}
}

func MaistraAdminRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"maistra.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
	}
}

func IstioAdminRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"config.istio.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"networking.istio.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"authentication.istio.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"rbac.istio.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
	}
}

func IstioRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{ 			
			APIGroups: []string{
				"",
			}, 			
			Resources: []string{
				"pods",
				"services",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
				"serviceaccounts",
				"namespaces",
				"routes",
			}, 			
			Verbs: []string{
				"*",
			 }, 
		}, 		
		{ 			
			APIGroups: []string{
		   		"apps",
			}, 			
			Resources: []string{
				"deployments",
				"daemonsets",
				"replicasets",
				"statefulsets",
		   	}, 			
			Verbs: []string{
				"*",
		 	}, 
		}, 		
		{ 			
			APIGroups: []string{
		   		"autoscaling",
		   	}, 			
			Resources: []string{
		   		"horizontalpodautoscalers",
		   	}, 			
			Verbs: []string{
				"*",
		 	}, 
		}, 		
		{ 			APIGroups: []string{
		   "extensions",
		   }, 			Resources: []string{
		   "daemonsets",
		   "deployments",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "batch",
		   }, 			Resources: []string{
		   "cronjobs",
		   "jobs",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "policy",
		   }, 			Resources: []string{
		   "poddisruptionbudgets",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "admissionregistration.k8s.io",
		   }, 			Resources: []string{
		   "mutatingwebhookconfigurations",
		   "validatingwebhookconfigurations",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "certmanager.k8s.io",
		   }, 			Resources: []string{
		   "clusterissuers",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "networking.k8s.io",
		   }, 			Resources: []string{
		   "networkpolicies",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		
		{ 			
			APIGroups: []string{
		   		"rbac.authorization.k8s.io",
		   	}, 			
			Resources: []string{
				"clusterrolebindings",
				"clusterroles",
				"roles",
				"rolebindings",
		   	}, 			
		   	Verbs: []string{
				"*",
		 	}, 
		}, 		
		{ 			APIGroups: []string{
		   "authentication.istio.io",
		   }, 			Resources: []string{
			"*",
		   "meshpolicies",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "config.istio.io",
		   }, 			Resources: []string{
			"*",
		   "attributemanifests",
		   "handlers",
		   "logentries",
		   "rules",
		   "metrics",
		   "kuberneteses",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "networking.istio.io",
		   }, 			Resources: []string{
		   
		   
		   
			"*",
		   "gateways",
		   "destinationrules",
		   "virtualservices",
		   "envoyfilters",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "monitoring.coreos.com",
		   }, 			Resources: []string{
		   "servicemonitors",
		   }, 			Verbs: []string{
		   "get",
		   "create",
		 }, }, 		{ 			APIGroups: []string{
		   "maistra.io",
		   }, 			Resources: []string{
		   
			"*",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "istio.openshift.com",
		   }, 			Resources: []string{
		   "controlplanes",
		   "controlplanes/status",
		   "controlplanes/finalizers",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "jaegertracing.io",
		   }, 			Resources: []string{
		   "jaegers",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "kiali.io",
		   }, 			Resources: []string{
		   "kialis",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "apps.openshift.io",
		   }, 			Resources: []string{
		   "deploymentconfigs",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "network.openshift.io",
		   }, 			Resources: []string{
		   "clusternetworks",
		   }, 			Verbs: []string{
		   "get",
		 }, }, 		{ 			APIGroups: []string{
		   "network.openshift.io",
		   }, 			Resources: []string{
		   "netnamespaces",
		   }, 			Verbs: []string{
		   "get",
		   "list",
		   "watch",
		   "update",
		 }, }, 		{ 			APIGroups: []string{
		   "oauth.openshift.io",
		   }, 			Resources: []string{
		   "oauthclients",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "project.openshift.io",
		   }, 			Resources: []string{
		   "projects",
		   "projectrequests",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "route.openshift.io",
		   }, 			Resources: []string{
		   "routes",
		   "routes/custom-host",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "security.openshift.io",
		   }, 			Resources: []string{
		   "securitycontextconstraints",
		   }, 			ResourceNames: []string{
		   "privileged",
		   "anyuid",
		   }, 			Verbs: []string{
		   "use",
		 
		 }, }, 		{ 			APIGroups: []string{
			"",
		   }, 			Resources: []string{
		   "nodes",
		   }, 			Verbs: []string{
		   "get",
		   "list",
		   "watch",
		 }, }, 		{ 			APIGroups: []string{
		   "extensions",
		   }, 			Resources: []string{
		   "ingresses",
		   }, 			Verbs: []string{
		   "get",
		   "list",
		   "watch",
		 }, }, 		{ 			APIGroups: []string{
		   "extensions",
		   "apps",
		   }, 			Resources: []string{
		   "deployments/finalizers",
		   }, 			ResourceNames: []string{
		   "istio-galley",
		   "istio-sidecar-injector",
		   }, 			Verbs: []string{
		   "update",
		 
		 }, }, 		{ 			APIGroups: []string{
		   "apiextensions.k8s.io",
		   }, 			Resources: []string{
		   "customresourcedefinitions",
		   }, 			Verbs: []string{
		   "get",
		   "list",
		   "watch",
		 }, }, 		{ 			APIGroups: []string{
		   "extensions",
		   }, 			Resources: []string{
		   "replicasets",
		   }, 			Verbs: []string{
		   "get",
		   "list",
		   "watch",
		 }, }, 		{ 			APIGroups: []string{
			"",
		   }, 			Resources: []string{
		   "replicationcontrollers",
		   }, 			Verbs: []string{
		   "get",
		   "list",
		   "watch",
		 
		 
		 }, }, 		{ 			APIGroups: []string{
		   "rbac.istio.io",
		   }, 			Resources: []string{
			"*",
		   }, 			Verbs: []string{
			"*",
		   "get",
		   "list",
		   "watch",
		 }, }, 		{ 			APIGroups: []string{
		   "apiextensions.k8s.io",
		   }, 			Resources: []string{
		   "customresourcedefinitions",
		   }, 			Verbs: []string{
			"*",
		 }, }, 		{ 			APIGroups: []string{
		   "extensions",
		   }, 			Resources: []string{
		   "ingresses",
		   "ingresses/status",
		   }, 			Verbs: []string{
			"*",
		 
		 }, }, 		{ 			APIGroups: []string{
			"",
		   }, 			Resources: []string{
		   "nodes/proxy",
		   }, 			Verbs: []string{
		   "get",
		   "list",
		   "watch",
		   },
		},
		{
		 NonResourceURLs: []string{
		   "/metrics",
		   }, 			
		Verbs: []string{
		   "get",
		 }, }, 		{ 			APIGroups: []string{
		   "authentication.k8s.io",
		   }, 			Resources: []string{
		   "tokenreviews",
		   }, 			Verbs: []string{
		   "create",
		 
		 }, }, 		{ 			APIGroups: []string{
		   "authorization.k8s.io",
		   }, 			Resources: []string{
		   "subjectaccessreviews",
		   }, 			Verbs: []string{
		   "create",
		 
		 
		 }, }, 		{ 			APIGroups: []string{
			 "k8s.cni.cncf.io",
		   }, 			Resources: []string{
			 "network-attachment-definitions",
		   }, 			Verbs: []string{
			 "create",
			 "delete",
			 "get",
			 "list",
			 "patch",
			 "watch",
		 }, 
		 },
	}
}

func IstioUserRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"networking.istio.io",
			},
			Resources: []string{
				"destinationrules",
				"gateways",
				"bypasses",
				"serviceentries",
				"sidecars",
				"virtualservices",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"authentication.istio.io",
			},
			Resources: []string{
				"policies",
				"meshpolicies",
			},
			Verbs: []string{
				"list",
			},
		},
	}
}