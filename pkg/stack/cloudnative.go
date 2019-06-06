package stack

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
)

func NewCloudNativeStack(cr *cloudnativev1alpha1.Workshop) *Stack {
	return &Stack{
		Name:        "Cloud-Native",
		Description: "Stack for Cloud-Native Development",
		Scope:       "general",
		WorkspaceConfig: WorkspaceConfig{
			Environments: Environments{
				Default: Environment{
					Recipe: Recipe{
						Type:    "dockerimage",
						Content: "docker-registry.default.svc:5000/openshift/che-cloud-native:latest",
					},
					Machines: Machines{
						DevMachine: DevMachine{
							Env: map[string]string{
								"MAVEN_OPTS":       "-Xmx512m",
								"MAVEN_MIRROR_URL": "http://nexus-repository.workshop-infra.svc:8081/repository/maven-all-public",
							},
							Servers: map[string]Server{
								"8080/tcp": Server{
									Attributes: map[string]string{},
									Protocol:   "http",
									Port:       "8080",
								},
								"8000/tcp": {
									Attributes: map[string]string{},
									Protocol:   "http",
									Port:       "8000",
								},
								"9000/tcp": {
									Attributes: map[string]string{},
									Protocol:   "http",
									Port:       "9000",
								},
								"9001/tcp": {
									Attributes: map[string]string{},
									Protocol:   "http",
									Port:       "9001",
								},
							},
							Volumes: map[string]string{},
							Installers: []string{
								"org.eclipse.che.exec",
								"org.eclipse.che.terminal",
								"org.eclipse.che.ws-agent",
								"org.eclipse.che.ls.java",
							},
							Attributes: map[string]string{
								"memoryLimitBytes": "2147483648",
							},
						},
					},
				},
			},
			Commands: []Command{
				{
					CommandLine: "mvn package -f ${current.project.path}",
					Name:        "build",
					Type:        "mvn",
					Attributes: map[string]string{
						"Goal":       "Build",
						"PreviewUrl": "",
					},
				},
				{
					CommandLine: "mvn clean package -f ${current.project.path}",
					Name:        "clean build",
					Type:        "mvn",
					Attributes: map[string]string{
						"Goal":       "Build",
						"PreviewUrl": "",
					},
				},
				{
					CommandLine: "mvn verify -f ${current.project.path}",
					Name:        "test",
					Type:        "mvn",
					Attributes: map[string]string{
						"Goal":       "Test",
						"PreviewUrl": "",
					},
				},
				{
					CommandLine: "mvn spring-boot:run -f ${current.project.path}",
					Name:        "spring-boot:run",
					Type:        "mvn",
					Attributes: map[string]string{
						"Goal":       "Run",
						"PreviewUrl": "${server.9000/tcp}",
					},
				},
				{
					CommandLine: "cd ${current.project.path} && java -Dswarm.http.port=9001 -jar target/*-thorntail.jar ",
					Name:        "thorntail:run",
					Type:        "mvn",
					Attributes: map[string]string{
						"Goal":       "Run",
						"PreviewUrl": "${server.9001/tcp}",
					},
				},
				{
					CommandLine: "mvn vertx:run -f ${current.project.path}",
					Name:        "vertx:run",
					Type:        "mvn",
					Attributes: map[string]string{
						"Goal":       "Run",
						"PreviewUrl": "${server.9001/tcp}",
					},
				},
				{
					CommandLine: "userName=${workspace.Namespace}\nprojectName=coolstore${userName#user}\nchmod +x /projects/labs/scripts/runGatewayService.sh\n/projects/labs/scripts/runGatewayService.sh ${projectName}",
					Name:        "runGatewayService",
					Attributes: map[string]string{
						"Goal":       "Run",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
				{
					CommandLine: "cd ${current.project.path} && mvn fabric8:deploy",
					Name:        "fabric8:deploy",
					Type:        "mvn",
					Attributes: map[string]string{
						"Goal":       "Deploy",
						"PreviewUrl": "",
					},
				},
				{
					CommandLine: "oc rollout pause dc/cart\noc set probe dc/cart --readiness --liveness --remove\noc patch dc/cart --patch '{\"spec\": {\"template\": {\"metadata\": {\"annotations\": {\"sidecar.istio.io/inject\": \"true\"}}}}}'\noc rollout resume dc/cart",
					Name:        "Cart Service: Inject Istio Sidecar",
					Attributes: map[string]string{
						"Goal":       "Service Mesh",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
				{
					CommandLine: "oc rollout pause dc/catalog\noc set probe dc/catalog --readiness --liveness --remove\noc patch dc/catalog --patch '{\"spec\": {\"template\": {\"metadata\": {\"annotations\": {\"sidecar.istio.io/inject\": \"true\"}}}}}'\noc patch dc/catalog --patch '{\"spec\": {\"template\": {\"spec\": {\"containers\": [{\"Name\": \"spring-boot\", \"command\": [\"/bin/bash\"], \"args\": [\"-c\", \"until $(curl -o /dev/null -s -I -f http://localhost:15000); do echo \\\"Waiting for Istio Sidecar...\\\"; sleep 1; done; sleep 10; /usr/local/s2i/run\"]}]}}}}'\noc rollout resume dc/catalog",
					Name:        "Catalog Service: Inject Istio Sidecar",
					Attributes: map[string]string{
						"Goal":       "Service Mesh",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
				{
					CommandLine: "oc rollout pause dc/inventory\noc set probe dc/inventory --readiness --liveness --remove\noc patch dc/inventory --patch '{\"spec\": {\"template\": {\"metadata\": {\"annotations\": {\"sidecar.istio.io/inject\": \"true\"}}}}}'\noc patch dc/inventory --patch '{\"spec\": {\"template\": {\"spec\": {\"containers\": [{\"Name\": \"thorntail-v2\", \"command\": [\"/bin/bash\"], \"args\": [\"-c\", \"until $(curl -o /dev/null -s -I -f http://localhost:15000); do echo \\\"Waiting for Istio Sidecar...\\\"; sleep 1; done; sleep 10; /usr/local/s2i/run\"]}]}}}}'\noc rollout resume dc/inventory",
					Name:        "Inventory Service: Inject Istio Sidecar",
					Attributes: map[string]string{
						"Goal":       "Service Mesh",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
				{
					CommandLine: "oc rollout pause dc/gateway\noc set probe dc/gateway --readiness --liveness --remove\noc patch dc/gateway --patch '{\"spec\": {\"template\": {\"metadata\": {\"annotations\": {\"sidecar.istio.io/inject\": \"true\"}}}}}'\noc patch dc/gateway --patch '{\"spec\": {\"template\": {\"spec\": {\"containers\": [{\"Name\": \"vertx\", \"command\": [\"/bin/bash\"], \"args\": [\"-c\", \"until $(curl -o /dev/null -s -I -f http://localhost:15000); do echo \\\"Waiting for Istio Sidecar...\\\"; sleep 1; done; sleep 10; /usr/local/s2i/run\"]}]}}}}'\noc rollout resume dc/gateway",
					Name:        "Gateway Service: Inject Istio Sidecar",
					Attributes: map[string]string{
						"Goal":       "Service Mesh",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
			},
			Projects:   []string{},
			DefaultEnv: "default",
			Name:       "default",
			Links:      []string{},
		},
		Components: []Component{
			{
				Version: "---",
				Name:    "CentOS",
			},
			{
				Version: "1.8.0_45",
				Name:    "JDK",
			},
			{
				Version: "3.2.2",
				Name:    "Maven",
			},
			{
				Version: "2.6",
				Name:    "Ansible",
			},
			{
				Version: "3.11",
				Name:    "OpenShift CLI",
			},
			{
				Version: "0.0.19",
				Name:    "OpenShift DO",
			},
		},
		Creator: "ide",
		Tags: []string{
			"Java",
			"JDK",
			"Maven",
			"CentOS",
			"Git",
		},
	}
}

func NewCloudNativeStackPermission(cr *cloudnativev1alpha1.Workshop, stackID string) *StackPermission {
	return &StackPermission{
		UserID:     "*",
		DomainID:   "stack",
		InstanceID: stackID,
		Actions: []string{
			"read",
			"search",
		},
	}
}
