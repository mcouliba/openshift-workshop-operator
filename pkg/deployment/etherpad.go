package deployment

import (
	"encoding/json"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type (
	EtherpadSettings struct {
		// Name your instance!
		Title string `json:"title"`
		// favicon default name
		// alternatively, set up a fully specified Url to your own favicon
		Favicon string `json:"favicon"`
		//IP and port which etherpad should bind at
		IP         string     `json:"ip"`
		Port       int        `json:"port"`
		DbType     string     `json:"dbType"`
		DbSettings DbSettings `json:"dbSettings"`
		//the default text of a pad
		DefaultPadText string `json:"defaultPadText"`
		/* Default Pad behavior, users can override by changing */
		PadOptions PadOptions `json:"padOptions"`
		/* Shoud we suppress errors from being visible in the default Pad Text? */
		SuppressErrorsInPadText bool `json:"suppressErrorsInPadText"`
		/* Users must have a session to access pads. This effectively allows only group pads to be accessed. */
		RequireSession bool `json:"requireSession"`
		/* Users may edit pads but not create new ones. Pad creation is only via the API. This applies both to group pads and regular pads. */
		EditOnly bool `json:"editOnly"`
		/* Users, who have a valid session, automatically get granted access to password protected pads */
		SessionNoPassword bool `json:"sessionNoPassword"`
		/* if true, all css & js will be minified before sending to the client. This will improve the loading performance massivly,
		   but makes it impossible to debug the javascript/css */
		Minify bool `json:"minify"`
		/* How long may clients use served javascript code (in seconds)? Without versioning this
		   may cause problems during deployment. Set to 0 to disable caching */
		MaxAge int `json:"maxAge"`
		/* This is the path to the Abiword executable. Setting it to null, disables abiword.
		   Abiword is needed to advanced import/export features of pads*/
		Abiword string `json:"abiword"`
		/* This is the path to the Tidy executable. Setting it to null, disables Tidy.
		   Tidy is used to improve the quality of exported pads*/
		TidyHTML string `json:"tidyHtml"`
		/* Allow import of file types other than the supported types */
		AllowUnknownFileEnds bool `json:"allowUnknownFileEnds"`
		/* This setting is used if you require authentication of all users.
		   Note: /admin always requires authentication. */
		RequireAuthentication bool `json:"requireAuthentication"`
		/* Require authorization by a module, or a user with is_admin set, see below. */
		RequireAuthorization bool `json:"requireAuthorization"`
		/*when you use NginX or another proxy/ load-balancer set this to true*/
		TrustProxy bool `json:"trustProxy"`
		/* Privacy= disable IP logging */
		DisableIPlogging bool `json:"disableIPlogging"`
		/* Users for basic authentication. is_admin = true gives access to /admin.
		   If you do not uncomment this, /admin will not be available! */
		Users Users `json:"users"`
		// restrict socket.io transport methods
		SocketTransportProtocols []string `json:"socketTransportProtocols"`
		// Allow Load Testing tools to hit the Etherpad Instance.  Warning this will disable security on the instance.
		LoadTest bool `json:"loadTest"`
		/* The log level we are using, can be: DEBUG, INFO, WARN, ERROR */
		Loglevel string `json:"loglevel"`
		//Logging configuration. See log4js documentation for further information
		// https://github.com/nomiddlename/log4js-node
		Logconfig Logconfig `json:"logconfig"`
	}

	DbSettings struct {
		User     string `json:"user"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Password string `json:"password"`
		Database string `json:"database"`
	}

	PadOptions struct {
		NoColors         bool   `json:"noColors"`
		ShowControls     bool   `json:"showControls"`
		ShowChat         bool   `json:"showChat"`
		ShowLineNumbers  bool   `json:"showLineNumbers"`
		UseMonospaceFont bool   `json:"useMonospaceFont"`
		UserName         bool   `json:"userName"`
		UserColor        bool   `json:"userColor"`
		Rtl              bool   `json:"rtl"`
		AlwaysShowChat   bool   `json:"alwaysShowChat"`
		ChatAndUsers     bool   `json:"chatAndUsers"`
		Lang             string `json:"lang"`
	}

	Users struct {
		Admin Admin `json:"admin"`
	}

	Admin struct {
		Password string `json:"password"`
		Is_admin bool   `json:"is_admin"`
	}

	Logconfig struct {
		Appenders []Appender `json:"appenders"`
	}

	Appender struct {
		Type string `json:"type"`
	}
)

func NewEtherpadSettingsJson(cr *openshiftv1alpha1.Workshop, defaultPadText string) string {

	settings := &EtherpadSettings{
		Title:   "OpenShift Workshop Etherpad",
		Favicon: "favicon.ico",
		IP:      "0.0.0.0",
		Port:    9001,
		DbType:  "mysql",
		DbSettings: DbSettings{
			User:     "DB_USER",
			Host:     "DB_HOST",
			Port:     "DB_PORT",
			Password: "DB_PASS",
			Database: "DB_DBID",
		},
		DefaultPadText: defaultPadText,
		PadOptions: PadOptions{
			NoColors:         false,
			ShowControls:     true,
			ShowChat:         true,
			ShowLineNumbers:  true,
			UseMonospaceFont: false,
			UserName:         false,
			UserColor:        false,
			Rtl:              false,
			AlwaysShowChat:   false,
			ChatAndUsers:     false,
			Lang:             "en-gb",
		},
		SuppressErrorsInPadText: false,
		RequireSession:          false,
		EditOnly:                false,
		SessionNoPassword:       false,
		Minify:                  true,
		MaxAge:                  21600, // 60 * 60 * 6 = 6 hours
		AllowUnknownFileEnds:    true,
		RequireAuthentication:   false,
		RequireAuthorization:    false,
		TrustProxy:              false,
		DisableIPlogging:        false,
		Users: Users{
			Admin: Admin{
				Password: "${MYSQL_ROOT_PASSWORD}",
				Is_admin: true,
			},
		},
		SocketTransportProtocols: []string{"xhr-polling", "jsonp-polling", "htmlfile"},
		LoadTest:                 false,
		Loglevel:                 "INFO",
		Logconfig: Logconfig{
			Appenders: []Appender{
				{
					Type: "console",
				},
			},
		},
	}

	jsonResult, _ := json.MarshalIndent(settings, "", "  ")
	return string(jsonResult)
}

func NewEtherpadDatabaseDeployment(cr *openshiftv1alpha1.Workshop, name string, namespace string) *appsv1.Deployment {
	etherpadDatabaseImage := "mysql:5.6"
	labels := GetLabels(cr, name)

	env := []corev1.EnvVar{
		{
			Name: "MYSQL_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "database-user",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
				},
			},
		},
		{
			Name: "MYSQL_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "database-password",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
				},
			},
		},
		{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "database-root-password",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
				},
			},
		},
		{
			Name: "MYSQL_DATABASE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "database-name",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
				},
			},
		},
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "mysql",
							Image:           etherpadDatabaseImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          name,
									ContainerPort: 3306,
									Protocol:      "TCP",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/sh",
											"-i",
											"-c",
											"MYSQL_PWD=\"$MYSQL_PASSWORD\" mysql -h 127.0.0.1 -u $MYSQL_USER -D $MYSQL_DATABASE -e 'SELECT 1'",
										},
									},
								},
								InitialDelaySeconds: 5,
								FailureThreshold:    10,
								TimeoutSeconds:      1,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      name + "-volume",
									MountPath: "/var/lib/mysql",
								},
							},
							Env: env,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: name + "-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: name,
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewEtherpadDeployment(cr *openshiftv1alpha1.Workshop, name string, namespace string) *appsv1.Deployment {
	etherpadImage := "quay.io/wkulhanek/etherpad:1.7.5"
	labels := GetLabels(cr, name)

	env := []corev1.EnvVar{
		{
			Name: "DB_DBID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "database-name",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name + "-mysql",
					},
				},
			},
		},
		{
			Name:  "DB_HOST",
			Value: name + "-mysql",
		},
		{
			Name: "DB_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "database-password",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name + "-mysql",
					},
				},
			},
		},
		{
			Name:  "DB_PORT",
			Value: "3306",
		},
		{
			Name: "DB_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "database-user",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name + "-mysql",
					},
				},
			},
		},
		{
			Name:  "NODE_ENV",
			Value: "production",
		},
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "mysql",
							Image:           etherpadImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9001,
									Protocol:      "TCP",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: int32(9001),
										},
										Scheme: corev1.URISchemeHTTP,
									},
								},
								FailureThreshold:    5,
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      1,
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: int32(9001),
										},
										Scheme: corev1.URISchemeHTTP,
									},
								},
								FailureThreshold:    3,
								InitialDelaySeconds: 120,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      1,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      name + "-settings",
									MountPath: "/opt/etherpad/config",
								},
							},
							Env: env,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: name + "-settings",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name + "-settings",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
