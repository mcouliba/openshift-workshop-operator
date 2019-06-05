package workshop

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	checlustercustomresource "github.com/redhat/cloud-native-workshop-operator/pkg/customresource/checluster"
	gogscustomresource "github.com/redhat/cloud-native-workshop-operator/pkg/customresource/gogs"
	nexuscustomresource "github.com/redhat/cloud-native-workshop-operator/pkg/customresource/nexus"
	deployment "github.com/redhat/cloud-native-workshop-operator/pkg/deployment"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_workshop")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Workshop Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWorkshop{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("workshop-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// register OpenShift routes in the scheme
	if err := routev1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// register Custom Resource Definition and Custom Resource in the scheme
	if err := apiextensionsv1beta1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	if err := nexuscustomresource.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	if err := gogscustomresource.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	if err := checlustercustomresource.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// Watch for changes to primary resource Workshop
	err = c.Watch(&source.Kind{Type: &cloudnativev1alpha1.Workshop{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Workshop
	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cloudnativev1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cloudnativev1alpha1.Workshop{},
	}); err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cloudnativev1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWorkshop implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWorkshop{}

// ReconcileWorkshop reconciles a Workshop object
type ReconcileWorkshop struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Workshop object and makes changes based on the state read
// and what is in the Workshop.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWorkshop) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Workshop")

	// Fetch the Workshop instance
	instance := &cloudnativev1alpha1.Workshop{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// // Define a new Pod object
	// pod := newPodForCR(instance)

	// // Set Workshop instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }
	//

	// list of the users endpoints for the workshop used in Etherpad
	var userEndpointStr strings.Builder

	// extract app route suffix from openshift-console
	openshiftConsoleRouteFound := &routev1.Route{}
	r.client.Get(context.TODO(), types.NamespacedName{Name: "console", Namespace: "openshift-console"}, openshiftConsoleRouteFound)
	re := regexp.MustCompile("^console.(.*?)$")
	match := re.FindStringSubmatch(openshiftConsoleRouteFound.Spec.Host)
	appsHostnameSuffix := match[1]

	enabledGuide := instance.Spec.Guide.Enabled

	// Setup for each user
	users := instance.Spec.Users
	if users < 0 {
		users = 0
	}

	id := 1

	for {
		username := fmt.Sprintf("user%d", id)
		coolstoreProjectName := fmt.Sprintf("coolstore%d", id)
		infraProjectName := fmt.Sprintf("infra%d", id)

		coolstoreNamespace := deployment.NewNamespace(instance, coolstoreProjectName)
		coolstoreNamespaceFound := &corev1.Namespace{}
		coolstoreNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: coolstoreNamespace.Name}, coolstoreNamespaceFound)

		if coolstoreNamespaceErr != nil && !errors.IsNotFound(coolstoreNamespaceErr) {
			reqLogger.Error(coolstoreNamespaceErr, "Failed to get Namespace", "Resource.name", coolstoreNamespace.Name)
			return reconcile.Result{}, coolstoreNamespaceErr
		}

		infraNamespace := deployment.NewNamespace(instance, infraProjectName)
		infraNamespaceFound := &corev1.Namespace{}
		infraNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: infraNamespace.Name}, infraNamespaceFound)

		if infraNamespaceErr != nil && !errors.IsNotFound(infraNamespaceErr) {
			reqLogger.Error(infraNamespaceErr, "Failed to get Namespace", "Resource.name", infraNamespace.Name)
			return reconcile.Result{}, infraNamespaceErr
		}

		if id <= users {
			// Add user endpoints in Etherpad
			userEndpointStr.WriteString(fmt.Sprintf("You are user%d\t|\thttp://guide-infra%d.%s\t|\t<INSERT_YOUR_NAME>\n", id, id, appsHostnameSuffix))

			// CoolStoreProject
			if coolstoreNamespaceErr != nil && errors.IsNotFound(coolstoreNamespaceErr) {
				reqLogger.Info("Creating a new Namespace", "Namespace.name", coolstoreNamespace.Name)
				if err := r.client.Create(context.TODO(), coolstoreNamespace); err != nil {
					reqLogger.Error(err, "Failed to create Namespace", "Resource.name", coolstoreNamespace.Name)
					return reconcile.Result{}, err
				}

				userRoleBinding := deployment.NewRoleBindingForUser(instance, fmt.Sprintf("user%d-admin", id), coolstoreNamespace.Name, username, "admin", "ClusterRole")
				if err := r.client.Create(context.TODO(), userRoleBinding); err != nil {
					return reconcile.Result{}, err
				}

				defaultRoleBinding := deployment.NewRoleBindingForServiceAccount(instance, "view", coolstoreNamespace.Name, "default", "view", "ClusterRole")
				if err := r.client.Create(context.TODO(), defaultRoleBinding); err != nil {
					return reconcile.Result{}, err
				}
			}

			// InfraProject
			if infraNamespaceErr != nil && errors.IsNotFound(infraNamespaceErr) {
				reqLogger.Info("Creating a new Namespace", "Namespace.name", infraNamespace.Name)
				if err := r.client.Create(context.TODO(), infraNamespace); err != nil {
					reqLogger.Error(err, "Failed to create Namespace", "Resource.name", infraNamespace.Name)
					return reconcile.Result{}, err
				}

				userRoleBinding := deployment.NewRoleBindingForUser(instance, username+"-admin", infraNamespace.Name, username, "admin", "ClusterRole")
				if err := r.client.Create(context.TODO(), userRoleBinding); err != nil {
					return reconcile.Result{}, err
				}
			}

			// Guide
			guideDeployment := deployment.NewWorkshopperDeployment(instance, "guide", infraNamespace.Name, coolstoreNamespace.Name, infraNamespace.Name, username, appsHostnameSuffix)
			guideDeploymentFound := &appsv1.Deployment{}
			guideDeploymentErr := r.client.Get(context.TODO(), types.NamespacedName{Name: guideDeployment.Name, Namespace: guideDeployment.Namespace}, guideDeploymentFound)

			guideRoute := deployment.NewRoute(instance, "guide", infraNamespace.Name, "guide", 8080)
			guideRouteFound := &routev1.Route{}
			guideRouteErr := r.client.Get(context.TODO(), types.NamespacedName{Name: guideRoute.Name, Namespace: guideRoute.Namespace}, guideRouteFound)

			guideService := deployment.NewService(instance, "guide", infraNamespace.Name, []string{"http"}, []int32{8080})
			guideServiceFound := &corev1.Service{}
			guideServiceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: guideService.Name, Namespace: guideService.Namespace}, guideServiceFound)

			if enabledGuide {
				if guideDeploymentErr != nil && errors.IsNotFound(guideDeploymentErr) {
					// Deploy Guide
					reqLogger.Info("Deploying a new Guide", "Namespace.name", infraNamespace.Name)
					if err := r.client.Create(context.TODO(), guideDeployment); err != nil {
						reqLogger.Error(err, "Failed to deploy Guide", "Resource.name", infraNamespace.Name)
						return reconcile.Result{}, err
					}
				} else {
					if !reflect.DeepEqual(guideDeployment.Spec.Template.Spec.Containers[0].Env, guideDeploymentFound.Spec.Template.Spec.Containers[0].Env) {
						// Update Guide
						reqLogger.Info("Updating the Guide", "Namespace.name", infraNamespace.Name)
						if err := r.client.Update(context.TODO(), guideDeployment); err != nil {
							reqLogger.Error(err, "Failed to update Guide", "Resource.name", infraNamespace.Name)
							return reconcile.Result{}, err
						}
					}
				}

				if guideServiceErr != nil && errors.IsNotFound(guideServiceErr) {
					// Create Service
					reqLogger.Info("Creating a Guide Service", "Namespace.name", infraNamespace.Name)
					if err := r.client.Create(context.TODO(), guideService); err != nil {
						reqLogger.Error(err, "Failed to create Guide Service", "Resource.name", infraNamespace.Name)
						return reconcile.Result{}, err
					}
				}

				if guideRouteErr != nil && errors.IsNotFound(guideRouteErr) {
					// Create Route
					reqLogger.Info("Creating a Guide Route", "Namespace.name", infraNamespace.Name)
					if err := r.client.Create(context.TODO(), guideRoute); err != nil {
						reqLogger.Error(err, "Failed to create Guide Route", "Resource.name", infraNamespace.Name)
						return reconcile.Result{}, err
					}
				}
			} else {
				if guideDeploymentErr == nil {
					// Undeploy Guide
					reqLogger.Info("Undeploying Guide")
					if err := r.client.Delete(context.TODO(), guideDeploymentFound); err != nil {
						reqLogger.Error(err, "Failed to undeploy Guide")
						return reconcile.Result{}, err
					}
				}

				if guideServiceErr == nil {
					// Delete Service
					reqLogger.Info("Deleting a Guide Service", "Namespace.name", infraNamespace.Name)
					if err := r.client.Delete(context.TODO(), guideServiceFound); err != nil {
						reqLogger.Error(err, "Failed to delete Guide Service", "Resource.name", infraNamespace.Name)
						return reconcile.Result{}, err
					}
				}

				if guideRouteErr == nil {
					// Delete Route
					reqLogger.Info("Deleting a Guide Route", "Namespace.name", infraNamespace.Name)
					if err := r.client.Delete(context.TODO(), guideRouteFound); err != nil {
						reqLogger.Error(err, "Failed to delete Guide Route", "Resource.name", infraNamespace.Name)
						return reconcile.Result{}, err
					}
				}
			}

		} else {
			if coolstoreNamespaceErr != nil && errors.IsNotFound(coolstoreNamespaceErr) && infraNamespaceErr != nil && errors.IsNotFound(infraNamespaceErr) {
				break
			}
			reqLogger.Info("Deleting a new Namespace", "Namespace.name", coolstoreNamespace.Name)
			if err := r.client.Delete(context.TODO(), coolstoreNamespace); err != nil {
				reqLogger.Error(err, "Failed to delete Namespace", "Resource.name", coolstoreNamespace.Name)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Deleting a new Namespace", "Namespace.name", infraNamespace.Name)
			if err := r.client.Delete(context.TODO(), infraNamespace); err != nil {
				reqLogger.Error(err, "Failed to delete Namespace", "Resource.name", infraNamespace.Name)
				return reconcile.Result{}, err
			}
		}

		id++
	}

	// Service Mesh
	enabledServiceMesh := instance.Spec.ServiceMesh.Enabled

	istioOperatorNamespace := deployment.NewNamespace(instance, "istio-operator")
	istioOperatorFound := &corev1.Namespace{}
	istioOperatorErr := r.client.Get(context.TODO(), types.NamespacedName{Name: istioOperatorNamespace.Name}, istioOperatorFound)

	if istioOperatorErr != nil && !errors.IsNotFound(istioOperatorErr) {
		reqLogger.Error(istioOperatorErr, "Failed to get istio-operator Namespace")
		return reconcile.Result{}, istioOperatorErr
	}

	istioSystemNamespace := deployment.NewNamespace(instance, "istio-system")
	istioSystemFound := &corev1.Namespace{}
	istioSystemErr := r.client.Get(context.TODO(), types.NamespacedName{Name: istioSystemNamespace.Name}, istioSystemFound)

	if istioSystemErr != nil && !errors.IsNotFound(istioSystemErr) {
		reqLogger.Error(istioSystemErr, "Failed to get istio-system Namespace")
		return reconcile.Result{}, istioSystemErr
	}

	if enabledServiceMesh {
		maistraVersion := instance.Spec.ServiceMesh.MaistraVersion
		if istioOperatorErr != nil && errors.IsNotFound(istioOperatorErr) {
			// Deploy istio-operator
			reqLogger.Info("Deploying istio-operator")
			if err := r.client.Create(context.TODO(), istioOperatorNamespace); err != nil {
				reqLogger.Error(err, "Failed to create Namespace", "Resource.name", istioOperatorNamespace.Name)
				return reconcile.Result{}, err
			}

			err = exec.Command("/bin/sh", "-c", "oc apply -n istio-operator -f https://raw.githubusercontent.com/Maistra/istio-operator/maistra-"+maistraVersion+"/deploy/servicemesh-operator.yaml").Run()
			if err != nil {
				reqLogger.Error(err, "Failed to deploy istio-operator")
				return reconcile.Result{}, err
			}
		}

		if istioSystemErr != nil && errors.IsNotFound(istioSystemErr) {
			// Deploy istio-system
			reqLogger.Info("Deploying istio-system")
			if err := r.client.Create(context.TODO(), istioSystemNamespace); err != nil {
				reqLogger.Error(err, "Failed to create Namespace", "Resource.name", istioSystemNamespace.Name)
				return reconcile.Result{}, err
			}

			err = exec.Command("/bin/sh", "-c", "oc delete limitrange --all -n istio-system;oc create -f https://raw.githubusercontent.com/Maistra/istio-operator/maistra-"+maistraVersion+"/deploy/examples/istio_v1alpha1_installation_cr.yaml").Run()
			if err != nil {
				reqLogger.Error(err, "Failed to deploy istio-system")
				return reconcile.Result{}, err
			}
		}
	}

	// Etherpad
	enabledEtherpad := instance.Spec.Etherpad.Enabled

	etherpadFound := &appsv1.Deployment{}
	etherpadErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "etherpad", Namespace: instance.Namespace}, etherpadFound)

	if etherpadErr != nil && !errors.IsNotFound(etherpadErr) {
		return reconcile.Result{}, etherpadErr
	}

	if enabledEtherpad {
		if etherpadErr != nil && errors.IsNotFound(etherpadErr) {

			reqLogger.Info("Creating Etherpad SQL Secret")
			databaseCredentials := map[string]string{
				"database-name":          "sampledb",
				"database-password":      "admin",
				"database-root-password": "admin",
				"database-user":          "admin",
			}
			etherpadDatabaseSecret := deployment.NewSecretStringData(instance, "etherpad-mysql", instance.Namespace, databaseCredentials)
			if err := r.client.Create(context.TODO(), etherpadDatabaseSecret); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Etherpad SQL Persistent Volume Claim")
			etherpadDatabasePersistentVolumeClaim := deployment.NewPersistentVolumeClaim(instance, "etherpad-mysql", instance.Namespace, "512Mi")
			if err := r.client.Create(context.TODO(), etherpadDatabasePersistentVolumeClaim); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Etherpad SQL Database")
			etherpadDatabaseDeployment := deployment.NewEtherpadDatabaseDeployment(instance, "etherpad-mysql", instance.Namespace)
			if err := r.client.Create(context.TODO(), etherpadDatabaseDeployment); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Etherpad SQL Service")
			etherpadDatabaseService := deployment.NewService(instance, "etherpad-mysql", instance.Namespace, []string{"mysql"}, []int32{3306})
			if err := r.client.Create(context.TODO(), etherpadDatabaseService); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Etherpad ConfigMap")
			defaultPadBytes, err := ioutil.ReadFile("template/etherpad.txt")
			if err != nil {
				return reconcile.Result{}, err
			}
			defaultPadText := strings.Replace(string(defaultPadBytes), "<USER_ENDPOINTS>", userEndpointStr.String(), 1)
			settings := map[string]string{
				"settings.json": deployment.NewEtherpadSettingsJson(instance, defaultPadText),
			}
			etherpadConfigMap := deployment.NewConfigMap(instance, "etherpad-settings", instance.Namespace, settings)
			if err := r.client.Create(context.TODO(), etherpadConfigMap); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Etherpad Deployment")
			etherpadDeployment := deployment.NewEtherpadDeployment(instance, "etherpad", instance.Namespace)
			if err := r.client.Create(context.TODO(), etherpadDeployment); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Etherpad Service")
			etherpadService := deployment.NewService(instance, "etherpad", instance.Namespace, []string{"http"}, []int32{9001})
			if err := r.client.Create(context.TODO(), etherpadService); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Etherpad Route")
			etherpadRoute := deployment.NewRoute(instance, "etherpad", instance.Namespace, "etherpad", 9001)
			if err := r.client.Create(context.TODO(), etherpadRoute); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
		}
	}

	// Nexus
	enabledNexus := instance.Spec.Nexus.Enabled

	nexusRepofound := &corev1.Pod{}
	nexusRepoErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "nexus-repository", Namespace: instance.Namespace}, nexusRepofound)

	if nexusRepoErr != nil && !errors.IsNotFound(nexusRepoErr) {
		return reconcile.Result{}, nexusRepoErr
	}

	if enabledNexus {
		if nexusRepoErr != nil && errors.IsNotFound(nexusRepoErr) {

			reqLogger.Info("Creating Nexus Custom Resource Definition")
			nexusCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "nexus.gpte.opentlc.com", "gpte.opentlc.com", "Nexus", "NexusList", "nexus", "nexus", "v1alpha1")
			if err := r.client.Create(context.TODO(), nexusCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Nexus Service Account")
			nexusServiceAccount := deployment.NewServiceAccount(instance, "nexus-operator", instance.Namespace)
			if err := r.client.Create(context.TODO(), nexusServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Nexus Cluster Role")
			nexusClusterRole := deployment.NewClusterRole(instance, "nexus-operator", instance.Namespace, deployment.NexusRules())
			if err := r.client.Create(context.TODO(), nexusClusterRole); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Nexus Cluster Role Binding")
			nexusClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "nexus-operator", instance.Namespace, "nexus-operator", "nexus-operator", "ClusterRole")
			if err := r.client.Create(context.TODO(), nexusClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Nexus Operator")
			nexusOperator := deployment.NewAnsibleOperatorDeployment(instance, "nexus-operator", instance.Namespace, "quay.io/wkulhanek/nexus-operator:v0.8.1", "nexus-operator")
			if err := r.client.Create(context.TODO(), nexusOperator); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Nexus Custom Resource")
			nexusCustomResource := deployment.NewNexusCustomResource(instance, "nexus-repository", instance.Namespace)
			if err := r.client.Create(context.TODO(), nexusCustomResource); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
		}
	}

	//Gogs
	enabledGogs := instance.Spec.Gogs.Enabled

	gogsPodfound := &corev1.Pod{}
	gogsPodErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "gogs", Namespace: instance.Namespace}, gogsPodfound)

	if gogsPodErr != nil && !errors.IsNotFound(gogsPodErr) {
		return reconcile.Result{}, gogsPodErr
	}

	if enabledGogs {
		if gogsPodErr != nil && errors.IsNotFound(gogsPodErr) {

			reqLogger.Info("Creating Gogs Custom Resource Definition")
			gogsCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "gogs.gpte.opentlc.com", "gpte.opentlc.com", "Gogs", "GogsList", "gogs", "gogs", "v1alpha1")
			if err := r.client.Create(context.TODO(), gogsCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Gogs Service Account")
			gogsServiceAccount := deployment.NewServiceAccount(instance, "gogs-operator", instance.Namespace)
			if err := r.client.Create(context.TODO(), gogsServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Gogs ClusterRole")
			gogsClusterRole := deployment.NewClusterRole(instance, "gogs-operator", instance.Namespace, deployment.GogsRules())
			if err := r.client.Create(context.TODO(), gogsClusterRole); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Gogs Cluster Role Binding")
			gogsClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "gogs-operator", instance.Namespace, "gogs-operator", "gogs-operator", "ClusterRole")
			if err := r.client.Create(context.TODO(), gogsClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Gogs Operator")
			gogsOperator := deployment.NewOperatorDeployment(instance, "gogs-operator", instance.Namespace, "quay.io/wkulhanek/gogs-operator:v0.0.6", "gogs-operator", nil)
			if err := r.client.Create(context.TODO(), gogsOperator); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Gogs Custom Resource")
			gogsCustomResource := deployment.NewGogsCustomResource(instance, "gogs-server", instance.Namespace)
			if err := r.client.Create(context.TODO(), gogsCustomResource); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
		}
	}

	// CodeReady Workspaces
	enabledWorkspaces := instance.Spec.Workspaces.Enabled
	enabledOpenShiftoAuth := instance.Spec.Workspaces.OpenShiftoAuth

	workspacesFound := &appsv1.Deployment{}
	workspacesErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "codeready", Namespace: "workspaces"}, workspacesFound)

	if workspacesErr != nil && !errors.IsNotFound(workspacesErr) {
		return reconcile.Result{}, workspacesErr
	}

	if enabledWorkspaces {
		if workspacesErr != nil && errors.IsNotFound(workspacesErr) {

			reqLogger.Info("Creating Workspaces Projects")
			workspacesNamespace := deployment.NewNamespace(instance, "workspaces")
			if err := r.client.Create(context.TODO(), workspacesNamespace); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Workspaces Custom Resource Definition")
			workspacesCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "checlusters.org.eclipse.che", "org.eclipse.che", "CheCluster", "CheClusterList", "checlusters", "checluster", "v1")
			if err := r.client.Create(context.TODO(), workspacesCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Workspaces Service Account")
			workspacesServiceAccount := deployment.NewServiceAccount(instance, "workspaces-operator", workspacesNamespace.Name)
			if err := r.client.Create(context.TODO(), workspacesServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Workspaces Cluster Role")
			workspacesClusterRole := deployment.NewClusterRole(instance, "workspaces-operator", workspacesNamespace.Name, deployment.WorkspacesRules())
			if err := r.client.Create(context.TODO(), workspacesClusterRole); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Workspaces Cluster Role Binding")
			workspacesClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "workspaces-operator", workspacesNamespace.Name, "workspaces-operator", "workspaces-operator", "ClusterRole")
			if err := r.client.Create(context.TODO(), workspacesClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Workspaces Operator")
			commands := []string{
				"che-operator",
			}
			workspacesOperator := deployment.NewOperatorDeployment(instance, "workspaces-operator", workspacesNamespace.Name, "registry.redhat.io/codeready-workspaces/server-operator-rhel8:1.2", "workspaces-operator", commands)
			if err := r.client.Create(context.TODO(), workspacesOperator); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Workspaces Custom Resource")
			workspacesCustomResource := deployment.NewCheClusterCustomResource(instance, "codeready", workspacesNamespace.Name, "registry.redhat.io/codeready-workspaces/server-rhel8", "1.2", false, false, enabledOpenShiftoAuth)
			if err := r.client.Create(context.TODO(), workspacesCustomResource); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}
