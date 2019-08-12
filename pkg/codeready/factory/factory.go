package codeready

func NewDebuggingFactory(openshiftConsoleURL string, openshiftAPIURl string, appsHostnameSuffix string, password string) *Factory {
	return &Factory{
		V:    "4.0",
		Name: "debugging-microservices",
		Creator: Creator{
			Name:  "Workshop Operator",
			Email: "workshop@operator.com",
		},
		Workspace: Workspace{
			Environments: Environments{
				Default: Environment{
					Recipe: Recipe{
						Type:    "dockerimage",
						Content: "image-registry.openshift-image-registry.svc:5000/openshift/che-cloud-native:ocp4",
					},
					Machines: Machines{
						DevMachine: DevMachine{
							Env: map[string]string{
								"MAVEN_OPTS":            "-Xmx512m",
								"MAVEN_MIRROR_URL":      "http://nexus-repository.workshop-infra.svc:8081/repository/maven-all-public",
								"OPENSHIFT_CONSOLE_URL": openshiftConsoleURL,
								"OPENSHIFT_API_URL":     openshiftAPIURl,
								"APPS_HOSTNAME_SUFFIX":  appsHostnameSuffix,
							},
							Servers: map[string]Server{},
							Volumes: map[string]string{},
							Installers: []string{
								"org.eclipse.che.exec",
								"org.eclipse.che.terminal",
								"org.eclipse.che.ws-agent",
								"org.eclipse.che.ls.java",
							},
							Attributes: map[string]string{
								"memoryLimitBytes": "4294967296",
							},
						},
					},
				},
			},
			Commands: []Command{
				{
					CommandLine: "username=${workspace.namespace}\nprojectname=coolstore${username#user}\noc login --server=${OPENSHIFT_API_URL}:6443 -u ${username} -p '" + password + "' --insecure-skip-tls-verify\noc project ${projectname}\necho \"-----------\"\necho \"Successful Connected to OpenShift as ${username}\"\necho \"-----------\"",
					Name:        "oc login",
					Attributes: map[string]string{
						"Goal":       "OpenShift",
						"PreviewUrl": "${OPENSHIFT_CONSOLE_URL}",
					},
					Type: "custom",
				},
				{
					CommandLine: "echo \"Squash version used from Solo.io\" && \\\nsquashctl --version",
					Name:        "Squash Version",
					Attributes: map[string]string{
						"Goal": "Debug",
					},
					Type: "custom",
				},
				{
					CommandLine: "oc start-build inventory-s2i --from-dir /projects/labs/inventory-thorntail/ --follow ",
					Name:        "Build Inventory Service",
					Attributes: map[string]string{
						"Goal": "Build",
					},
					Type: "custom",
				},
				{
					CommandLine: "oc new-app /projects/labs/ --strategy=docker --context-dir=catalog-go --name=catalog-v2 --labels app=catalog,group=com.redhat.cloudnative,provider=fabric8,version=2.0 && \\\noc start-build catalog-v2 --from-dir /projects/labs/catalog-go/ --follow && \\\noc patch dc/catalog-v2 --patch '{\"spec\": {\"template\": {\"metadata\": {\"annotations\": {\"sidecar.istio.io/inject\": \"true\"}}}}}' && \\\necho \"Catalog V2 in Go Language and Integration of Istio Sidecar\"",
					Name:        "Option: Catalog V2  in Go",
					Attributes: map[string]string{
						"Goal": "Build",
					},
					Type: "custom",
				},
				{
					CommandLine: "username=${workspace.namespace}\nprojectname=coolstore${username#user}\noc process -f /projects/labs/openshift/mysteriousapp.yml -p COOLSTORE_PROJECT=${projectname} -p APPS_HOSTNAME_SUFFIX=${APPS_HOSTNAME_SUFFIX} | oc create -f - && \\\noc start-build catalog-s2i --from-dir /projects/labs/catalog-spring-boot/ && \\\noc start-build inventory-s2i --from-dir /projects/labs/inventory-thorntail/ --follow && \\\noc start-build gateway-s2i --from-dir /projects/labs/gateway-vertx/ && \\\noc start-build web --from-dir /projects/labs/web-nodejs/",
					Name:        "Build Mysterious Application",
					Attributes: map[string]string{
						"Goal": "Build",
					},
					Type: "custom",
				},
				{
					CommandLine: "username=${workspace.namespace}\nprojectname=coolstore${username#user}\ninfraname=infra${username#user}\necho \"Execute in Terminal Windows\" && \\\necho \"squashctl --namespace ${projectname} --debugger java-port --squash-namespace ${infraname}\"",
					Name:        "Debug Squash Inventory",
					Attributes: map[string]string{
						"Goal": "Debug",
					},
					Type: "custom",
				},
				{
					CommandLine: "username=${workspace.namespace}\nprojectname=infra${username#user}\necho \"squashctl utils delete-planks --squash-namespace ${projectname}\"\nsquashctl utils delete-planks --squash-namespace ${projectname}",
					Name:        "Delete Squash Attachments",
					Attributes: map[string]string{
						"Goal": "Debug",
					},
					Type: "custom",
				},
				{
					CommandLine: "username=${workspace.namespace}\nproject_name=coolstore${username#user}\nurl=http://istio-ingressgateway.istio-system.svc/${project_name}/api/products\necho \"$url\"\nresponse=$(curl --write-out %{http_code} --silent --output /dev/null $url)\nif [ \"${response}\" == \"200\" ]\nthen\n    while true; do \n        if curl -s ${url} | grep -q OFFICIAL\n        then\n            echo \"Gateway => Catalog GoLang (v2)\";\n        else\n            echo \"Gateway => Catalog Spring Boot (v1)\";\n        fi\n        sleep 1\n    done\nelse\n    echo \"Error ${response} when calling ${url}\";\nfi",
					Name:        "testGatewayService",
					Attributes: map[string]string{
						"Goal":       "Test",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
				{
					CommandLine: "username=${workspace.namespace}\nproject_name=coolstore${username#user}\nurl=http://istio-ingressgateway.istio-system.svc/${project_name}/api/products\necho \"$url\"\nresponse=$(curl --write-out %{http_code} --silent --output /dev/null $url)\nif [ \"${response}\" == \"200\" ]\nthen\n    while true; do \n        if curl -s ${url} | grep -q OFFICIAL\n        then\n            echo \"Gateway => Catalog GoLang (v2)\";\n        else\n            echo \"Gateway => Catalog Spring Boot (v1)\";\n        fi\n        sleep 1\n    done\nelse\n    echo \"Error ${response} when calling ${url}\";\nfi",
					Name:        "testGatewayService",
					Attributes: map[string]string{
						"Goal":       "Test",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
				{
					CommandLine: "username=${workspace.namespace}\nproject_name=coolstore${username#user}\nurl=http://istio-ingressgateway.istio-system.svc/${project_name}/api/products\necho \"$url\"\nresponse=$(curl --write-out %{http_code} --silent --output /dev/null $url)\nif [ \"${response}\" == \"200\" ]\nthen\n    while true; do \n        if curl -s ${url} | grep -q OFFICIAL\n        then\n            echo \"Gateway => Catalog GoLang (v2)\";\n        else\n            echo \"Gateway => Catalog Spring Boot (v1)\";\n        fi\n        sleep 1\n    done\nelse\n    echo \"Error ${response} when calling ${url}\";\nfi",
					Name:        "testGatewayService",
					Attributes: map[string]string{
						"Goal":       "Test",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
				{
					CommandLine: "username=${workspace.namespace}\nproject_name=coolstore${username#user}\nurl=http://web.${project_name}.svc.cluster.local:8080\necho \"$url\"\nresponse=$(curl --write-out %{http_code} --silent --output /dev/null $url)\nif [ \"${response}\" == \"200\" ]\nthen\n    while true; do \n        echo \"Generating traffic\";\n        sleep 1\n    done\nelse\n    echo \"Error ${response} when calling ${url}\";\nfi",
					Name:        "generateTraffic",
					Attributes: map[string]string{
						"Goal":       "Test",
						"PreviewUrl": "",
					},
					Type: "custom",
				},
			},
			Projects: []Project{
				{
					Name: "labs",
					Type: "blank",
					Source: Source{
						Location: "https://github.com/mcouliba/cloud-native-labs/archive/debugging.zip",
						Type:     "zip",
						Parameters: map[string]string{
							"skipFirstLevel": "true",
						},
					},
					Path: "/labs",
				},
				{
					Name: "gateway-vertx",
					Type: "maven",
					Source: Source{
						Parameters: map[string]string{},
					},
					Path: "/labs/gateway-vertx",
				},
				{
					Name: "inventory-thorntail",
					Type: "maven",
					Source: Source{
						Parameters: map[string]string{},
					},
					Path: "/labs/inventory-thorntail",
				},
				{
					Name: "cart-quarkus",
					Type: "maven",
					Source: Source{
						Parameters: map[string]string{},
					},
					Path: "/labs/cart-quarkus",
				},
				{
					Name: "catalog-spring-boot",
					Type: "maven",
					Source: Source{
						Parameters: map[string]string{},
					},
					Path: "/labs/catalog-spring-boot",
				},
			},
			DefaultEnv: "default",
			Name:       "wksp-debugging",
			Links:      []string{},
		},
	}
}
