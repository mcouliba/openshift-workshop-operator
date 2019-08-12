package workshop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	codereadyfactory "github.com/redhat/openshift-workshop-operator/pkg/codeready/factory"
	codereadystack "github.com/redhat/openshift-workshop-operator/pkg/codeready/stack"
	codereadyuser "github.com/redhat/openshift-workshop-operator/pkg/codeready/user"
	checlustercustomresource "github.com/redhat/openshift-workshop-operator/pkg/customresource/checluster"
	gogscustomresource "github.com/redhat/openshift-workshop-operator/pkg/customresource/gogs"
	nexuscustomresource "github.com/redhat/openshift-workshop-operator/pkg/customresource/nexus"
	"github.com/redhat/openshift-workshop-operator/pkg/deployment"
	smcp "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshcontrolplane"
	smmr "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshmemberroll"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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

	// register OpenShift Routes in the scheme
	if err := routev1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// register OpenShift Image in the scheme
	if err := imagev1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// register OpenShift Security in the scheme
	if err := securityv1.AddToScheme(mgr.GetScheme()); err != nil {
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

	if err := smcp.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := smmr.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// Watch for changes to primary resource Workshop
	err = c.Watch(&source.Kind{Type: &openshiftv1alpha1.Workshop{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Workshop
	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	}); err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
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
	instance := &openshiftv1alpha1.Workshop{}
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

	var (
		userEndpointStr     strings.Builder
		openshiftConsoleURL string
		openshiftAPIURL     string
		appsHostnameSuffix  string
	)
	// extract app route suffix from openshift-console
	openshiftConsoleRouteFound := &routev1.Route{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "console", Namespace: "openshift-console"}, openshiftConsoleRouteFound); err != nil {
		reqLogger.Error(err, "Failed to get OpenShift Console")
		return reconcile.Result{}, err
	}
	openshiftConsoleURL = "https://" + openshiftConsoleRouteFound.Spec.Host

	re := regexp.MustCompile("^console-openshift-console\\.apps\\.(.*?)$")
	match := re.FindStringSubmatch(openshiftConsoleRouteFound.Spec.Host)
	openshiftAPIURL = "https://api." + match[1]

	re = regexp.MustCompile("^console-openshift-console\\.(.*?)$")
	match = re.FindStringSubmatch(openshiftConsoleRouteFound.Spec.Host)
	appsHostnameSuffix = match[1]

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
				if err := r.client.Create(context.TODO(), coolstoreNamespace); err != nil && !errors.IsAlreadyExists(err) {
					reqLogger.Error(err, "Failed to create Namespace", "Resource.name", coolstoreNamespace.Name)
					return reconcile.Result{}, err
				} else if err == nil {
					reqLogger.Info("Created " + coolstoreNamespace.Name + " project")
				}

				istioRole := deployment.NewRole(deployment.NewRoleParameters{
					Name:      username + "-istio",
					Namespace: coolstoreNamespace.Name,
					Rules:     deployment.IstioUserRules(),
				})
				if err := r.client.Create(context.TODO(), istioRole); err != nil && !errors.IsAlreadyExists(err) {
					return reconcile.Result{}, err
				} else if err == nil {
					reqLogger.Info("Created " + username + "-istio" + " Role")
				}

				istioRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
					Name:      username + "-istio",
					Namespace: coolstoreNamespace.Name,
					Username:  username,
					RoleName:  username + "-istio",
					RoleKind:  "Role",
				})
				if err := r.client.Create(context.TODO(), istioRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
					return reconcile.Result{}, err
				} else if err == nil {
					reqLogger.Info("Created " + username + "-istio" + " Role Binding")
				}

				userRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
					Name:      username + "-admin",
					Namespace: coolstoreNamespace.Name,
					Username:  username,
					RoleName:  "admin",
					RoleKind:  "ClusterRole",
				})
				if err := r.client.Create(context.TODO(), userRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
					return reconcile.Result{}, err
				} else if err == nil {
					reqLogger.Info("Created " + username + "-admin" + " Role Binding")
				}

				defaultRoleBinding := deployment.NewRoleBindingSA(deployment.NewRoleBindingSAParameters{
					Name:               "view",
					Namespace:          coolstoreNamespace.Name,
					ServiceAccountName: "default",
					RoleName:           "view",
					RoleKind:           "ClusterRole",
				})
				if err := r.client.Create(context.TODO(), defaultRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
					return reconcile.Result{}, err
				} else if err == nil {
					reqLogger.Info("Created View Role Binding")
				}
			}

			//SCC
			serviceaccount := "system:serviceaccount:" + coolstoreProjectName + ":default"

			privilegedSCCFound := &securityv1.SecurityContextConstraints{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "privileged"}, privilegedSCCFound); err != nil {
				reqLogger.Error(err, "Failed to get Privileged SCC", "Resource.name", "privileged")
				return reconcile.Result{}, err
			}

			if !util.StringInSlice(serviceaccount, privilegedSCCFound.Users) {
				privilegedSCCFound.Users = append(privilegedSCCFound.Users, serviceaccount)
				reqLogger.Info("Updating the Privileged SCC")
				if err := r.client.Update(context.TODO(), privilegedSCCFound); err != nil {
					reqLogger.Error(err, "Failed to update the Privileged SCC")
					return reconcile.Result{}, err
				}
			}

			anyuidSCCFound := &securityv1.SecurityContextConstraints{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "anyuid"}, anyuidSCCFound); err != nil {
				reqLogger.Error(err, "Failed to get Anyuid SCC", "Resource.name", "anyuid")
				return reconcile.Result{}, err
			}

			if !util.StringInSlice(serviceaccount, anyuidSCCFound.Users) {
				anyuidSCCFound.Users = append(anyuidSCCFound.Users, serviceaccount)
				reqLogger.Info("Updating the Anyuid SCC")
				if err := r.client.Update(context.TODO(), anyuidSCCFound); err != nil {
					reqLogger.Error(err, "Failed to update the Anyuid SCC")
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

				userRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
					Name:      username + "-admin",
					Namespace: infraNamespace.Name,
					Username:  username,
					RoleName:  "admin",
					RoleKind:  "ClusterRole",
				})
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

	//////////////////////////
	// Service Mesh
	//////////////////////////
	enabledServiceMesh := instance.Spec.ServiceMesh.Enabled
	jaegerOperatorImage := instance.Spec.ServiceMesh.JaegerOperatorImage
	kialiOperatorImage := instance.Spec.ServiceMesh.KialiOperatorImage
	istioOperatorImage := instance.Spec.ServiceMesh.IstioOperatorImage

	jaegerOperatorNamespace := deployment.NewNamespace(instance, "observability")
	jaegerOperatorFound := &corev1.Namespace{}
	jaegerOperatorErr := r.client.Get(context.TODO(), types.NamespacedName{Name: jaegerOperatorNamespace.Name}, jaegerOperatorFound)

	if jaegerOperatorErr != nil && !errors.IsNotFound(jaegerOperatorErr) {
		reqLogger.Error(jaegerOperatorErr, "Failed to get observability Namespace")
		return reconcile.Result{}, jaegerOperatorErr
	}

	kialiOperatorNamespace := deployment.NewNamespace(instance, "kiali-operator")
	kialiOperatorFound := &corev1.Namespace{}
	kialiOperatorErr := r.client.Get(context.TODO(), types.NamespacedName{Name: kialiOperatorNamespace.Name}, kialiOperatorFound)

	if kialiOperatorErr != nil && !errors.IsNotFound(kialiOperatorErr) {
		reqLogger.Error(kialiOperatorErr, "Failed to get observability Namespace")
		return reconcile.Result{}, kialiOperatorErr
	}

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
		// Jaeger
		if jaegerOperatorErr != nil && errors.IsNotFound(jaegerOperatorErr) {
			reqLogger.Info("Creating Jaeger Project")
			if err := r.client.Create(context.TODO(), jaegerOperatorNamespace); err != nil {
				reqLogger.Error(err, "Failed to create Namespace", "Resource.name", jaegerOperatorNamespace.Name)
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Jaeger Custom Resource Definition")
			jaegerOperatorCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "jaegers.jaegertracing.io", "jaegertracing.io", "Jaeger", "JaegerList", "jaegers", "jaeger", "v1", nil, nil)
			if err := r.client.Create(context.TODO(), jaegerOperatorCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Jaeger Service Account")
			jaegerOperatorServiceAccount := deployment.NewServiceAccount(instance, "jaeger-operator", jaegerOperatorNamespace.Name)
			if err := r.client.Create(context.TODO(), jaegerOperatorServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Jaeger Cluster Role")
			jaegerOperatorClusterRole := deployment.NewClusterRole(instance, "jaeger-operator", jaegerOperatorNamespace.Name, deployment.JaegerRules())
			if err := r.client.Create(context.TODO(), jaegerOperatorClusterRole); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Jaeger Cluster Role Binding")
			jaegerOperatorClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "jaeger-operator", jaegerOperatorNamespace.Name, "jaeger-operator", "jaeger-operator", "ClusterRole")
			if err := r.client.Create(context.TODO(), jaegerOperatorClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			args := []string{
				"start",
			}
			reqLogger.Info("Creating Jaeger Operator")
			jaegerOperator := deployment.NewOperatorDeployment(instance, "jaeger-operator", jaegerOperatorNamespace.Name, jaegerOperatorImage, "jaeger-operator", 8383, nil, args, nil, nil)
			if err := r.client.Create(context.TODO(), jaegerOperator); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
		}

		// Kiali
		if kialiOperatorErr != nil && errors.IsNotFound(kialiOperatorErr) {
			reqLogger.Info("Creating Kiali Project")
			if err := r.client.Create(context.TODO(), kialiOperatorNamespace); err != nil {
				reqLogger.Error(err, "Failed to create Namespace", "Resource.name", kialiOperatorNamespace.Name)
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Kiali Custom Resource Definition")
			kialiOperatorCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "kialis.kiali.io", "kiali.io", "Kiali", "KialiList", "kialis", "kiali", "v1alpha1", nil, nil)
			if err := r.client.Create(context.TODO(), kialiOperatorCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
			kialiOperatorCustomResourceDefinition2 := deployment.NewCustomResourceDefinition(instance, "monitoringdashboards.monitoring.kiali.io", "monitoring.kiali.io", "MonitoringDashboard", "MonitoringDashboardList", "monitoringdashboards", "monitoringdashboard", "v1alpha1", nil, nil)
			if err := r.client.Create(context.TODO(), kialiOperatorCustomResourceDefinition2); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Kiali Service Account")
			kialiOperatorServiceAccount := deployment.NewServiceAccount(instance, "kiali-operator", kialiOperatorNamespace.Name)
			if err := r.client.Create(context.TODO(), kialiOperatorServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Kiali Cluster Role")
			kialiOperatorClusterRole := deployment.NewClusterRole(instance, "kiali-operator", kialiOperatorNamespace.Name, deployment.KialiRules())
			if err := r.client.Create(context.TODO(), kialiOperatorClusterRole); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Kiali Cluster Role Binding")
			kialiOperatorClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "kiali-operator", kialiOperatorNamespace.Name, "kiali-operator", "kiali-operator", "ClusterRole")
			if err := r.client.Create(context.TODO(), kialiOperatorClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Kiali Operator")
			kialiOperator := deployment.NewAnsibleOperatorDeployment(instance, "kiali-operator", kialiOperatorNamespace.Name, kialiOperatorImage, "kiali-operator")
			if err := r.client.Create(context.TODO(), kialiOperator); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
		}

		// ISTIO
		if istioOperatorErr != nil && errors.IsNotFound(istioOperatorErr) {
			// Deploy istio-operator
			reqLogger.Info("Creating istio-system Project")
			if err := r.client.Create(context.TODO(), istioOperatorNamespace); err != nil {
				reqLogger.Error(err, "Failed to create Namespace", "Resource.name", istioOperatorNamespace.Name)
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Istio Custom Resource Definition")
			istioOperatorCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "controlplanes.istio.openshift.com", "istio.openshift.com", "ControlPlane", "ControlPlaneList", "controlplanes", "controlplane", "v1alpha3", nil, nil)
			if err := r.client.Create(context.TODO(), istioOperatorCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
			istioOperatorCustomResourceDefinition2 := deployment.NewCustomResourceDefinition(instance, "servicemeshcontrolplanes.maistra.io", "maistra.io", "ServiceMeshControlPlane", "ServiceMeshControlPlaneList", "servicemeshcontrolplanes", "servicemeshcontrolplane", "v1", []string{"smcp"}, nil)
			if err := r.client.Create(context.TODO(), istioOperatorCustomResourceDefinition2); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			additionalPrinterColumns := []apiextensionsv1beta1.CustomResourceColumnDefinition{
				{
					JSONPath:    ".spec.members",
					Description: "Namespaces that are members of this Control Plane",
					Name:        "Members",
					Type:        "string",
				},
			}
			istioOperatorCustomResourceDefinition3 := deployment.NewCustomResourceDefinition(instance, "servicemeshmemberrolls.maistra.io", "maistra.io", "ServiceMeshMemberRoll", "ServiceMeshMemberRollList", "servicemeshmemberrolls", "servicemeshmemberroll", "v1", []string{"smmr"}, additionalPrinterColumns)
			if err := r.client.Create(context.TODO(), istioOperatorCustomResourceDefinition3); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Istio Service Account")
			istioOperatorServiceAccount := deployment.NewServiceAccount(instance, "istio-operator", istioOperatorNamespace.Name)
			if err := r.client.Create(context.TODO(), istioOperatorServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Istio Cluster Role")
			istioOperatorClusterRole := deployment.NewClusterRole(instance, "maistra-admin", istioOperatorNamespace.Name, deployment.MaistraAdminRules())
			if err := r.client.Create(context.TODO(), istioOperatorClusterRole); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
			istioOperatorClusterRole2 := deployment.NewClusterRole(instance, "istio-admin", istioOperatorNamespace.Name, deployment.IstioAdminRules())
			if err := r.client.Create(context.TODO(), istioOperatorClusterRole2); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
			istioOperatorClusterRole3 := deployment.NewClusterRole(instance, "istio-operator", istioOperatorNamespace.Name, deployment.IstioRules())
			if err := r.client.Create(context.TODO(), istioOperatorClusterRole3); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Istio Cluster Role Binding")
			istioOperatorClusterRoleBinding := deployment.NewClusterRoleBinding(instance, "maistra-admin", istioOperatorNamespace.Name, "maistra-admin", "ClusterRole")
			if err := r.client.Create(context.TODO(), istioOperatorClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
			istioOperatorClusterRoleBinding2 := deployment.NewClusterRoleBinding(instance, "istio-admin", istioOperatorNamespace.Name, "istio-admin", "ClusterRole")
			if err := r.client.Create(context.TODO(), istioOperatorClusterRoleBinding2); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
			istioOperatorClusterRoleBinding3 := deployment.NewClusterRoleBindingForServiceAccount(instance, "istio-operator", istioOperatorNamespace.Name, "istio-operator", "istio-operator", "ClusterRole")
			if err := r.client.Create(context.TODO(), istioOperatorClusterRoleBinding3); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			istioValidatingWebhookConfiguration := deployment.NewValidatingWebhookConfiguration(instance, "istio-operator.servicemesh-resources.maistra.io", deployment.IstioWebHook())
			if err := r.client.Create(context.TODO(), istioValidatingWebhookConfiguration); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			targetPort := intstr.IntOrString{
				IntVal: 11999,
			}
			reqLogger.Info("Creating Admin Controller Service")
			adminControllerService := deployment.NewCustomService(instance, "admission-controller", instance.Namespace, []string{"admin"}, []int32{443}, []intstr.IntOrString{targetPort})
			if err := r.client.Create(context.TODO(), adminControllerService); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			commands := []string{
				"istio-operator",
				"--discoveryCacheDir",
				"/home/istio-operator/.kube/cache/discovery",
			}
			volumeMounts := []corev1.VolumeMount{
				{
					Name:      "discovery-cache",
					MountPath: "/home/istio-operator/.kube/cache/discovery",
				},
			}
			volumes := []corev1.Volume{
				{
					Name: "discovery-cache",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			}
			reqLogger.Info("Creating Istio Operator")
			istioOperator := deployment.NewOperatorDeployment(instance, "istio-operator", istioOperatorNamespace.Name, istioOperatorImage, "istio-operator", 60000, commands, nil, volumeMounts, volumes)
			if err := r.client.Create(context.TODO(), istioOperator); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

		}

		// Wait for Istio Operator to be running
		time.Sleep(30 * time.Second)

		// ISTIO-SYSTEM
		if istioSystemErr != nil && errors.IsNotFound(istioSystemErr) {
			// Deploy istio-system
			if err := r.client.Create(context.TODO(), istioSystemNamespace); err != nil {
				reqLogger.Error(err, "Failed to create Namespace", "Resource.name", istioSystemNamespace.Name)
				return reconcile.Result{}, err
			} else if err == nil {
				reqLogger.Info("Created istio-system Project")
			}

			serviceMeshControlPlaneCR := smcp.NewServiceMeshControlPlaneCR(smcp.NewServiceMeshControlPlaneCRParameters{
				Name:      "full-install",
				Namespace: istioSystemNamespace.Name,
			})
			if err := r.client.Create(context.TODO(), serviceMeshControlPlaneCR); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			} else if err == nil {
				reqLogger.Info("Created ServiceMeshControlPlane Custom Resource")
			}

			istioMembers := []string{}
			for id := 1; id <= users; id++ {
				istioMembers = append(istioMembers, fmt.Sprintf("coolstore%d", id))
			}

			serviceMeshMemberRollCR := smmr.NewServiceMeshMemberRollCR(smmr.NewServiceMeshMemberRollCRParameters{
				Name:      "default",
				Namespace: istioSystemNamespace.Name,
				Members:   istioMembers,
			})
			if err := r.client.Create(context.TODO(), serviceMeshMemberRollCR); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			} else if err == nil {
				reqLogger.Info("Created ServiceMeshControlPlane Custom Resource")
			}
		}

	}

	// Etherpad
	if err := r.reconcileEtherpad(instance, userEndpointStr); err != nil {
		return reconcile.Result{}, err
	}

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

	//////////////////////////
	// Nexus
	//////////////////////////
	enabledNexus := instance.Spec.Nexus.Enabled

	if enabledNexus {
		nexusNamespace := deployment.NewNamespace(instance, "opentlc-shared")
		if err := r.client.Create(context.TODO(), nexusNamespace); err != nil && !errors.IsAlreadyExists(err) {
			reqLogger.Error(err, "Failed to create Namespace", "Resource.name", nexusNamespace.Name)
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Nexus Project")
		}

		nexusCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "nexus.gpte.opentlc.com", "gpte.opentlc.com", "Nexus", "NexusList", "nexus", "nexus", "v1alpha1", nil, nil)
		if err := r.client.Create(context.TODO(), nexusCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created  Nexus Custom Resource Definition")
		}

		nexusServiceAccount := deployment.NewServiceAccount(instance, "nexus-operator", nexusNamespace.Name)
		if err := r.client.Create(context.TODO(), nexusServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created  Nexus Service Account")
		}

		nexusClusterRole := deployment.NewClusterRole(instance, "nexus-operator", nexusNamespace.Name, deployment.NexusRules())
		if err := r.client.Create(context.TODO(), nexusClusterRole); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created  Nexus Cluster Role")
		}

		nexusClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "nexus-operator", nexusNamespace.Name, "nexus-operator", "nexus-operator", "ClusterRole")
		if err := r.client.Create(context.TODO(), nexusClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Nexus Cluster Role Binding")
		}

		nexusOperator := deployment.NewAnsibleOperatorDeployment(instance, "nexus-operator", nexusNamespace.Name, "quay.io/gpte-devops-automation/nexus-operator:v0.9", "nexus-operator")
		if err := r.client.Create(context.TODO(), nexusOperator); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Nexus Operator")
		}

		nexusCustomResource := nexuscustomresource.NewNexusCustomResource(instance, "nexus", nexusNamespace.Name)
		if err := r.client.Create(context.TODO(), nexusCustomResource); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Nexus Custom Resource")
		}
	}

	//////////////////////////
	// Gogs
	//////////////////////////
	enabledGogs := instance.Spec.Gogs.Enabled

	gogsPodfound := &corev1.Pod{}
	gogsPodErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "gogs-gogs-server", Namespace: instance.Namespace}, gogsPodfound)

	if gogsPodErr != nil && !errors.IsNotFound(gogsPodErr) {
		return reconcile.Result{}, gogsPodErr
	}

	if enabledGogs {
		if gogsPodErr != nil && errors.IsNotFound(gogsPodErr) {

			reqLogger.Info("Creating Gogs Custom Resource Definition")
			gogsCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "gogs.gpte.opentlc.com", "gpte.opentlc.com", "Gogs", "GogsList", "gogs", "gogs", "v1alpha1", nil, nil)
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
			gogsOperator := deployment.NewOperatorDeployment(instance, "gogs-operator", instance.Namespace, "quay.io/wkulhanek/gogs-operator:v0.0.6", "gogs-operator", 60000, nil, nil, nil, nil)
			if err := r.client.Create(context.TODO(), gogsOperator); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating Gogs Custom Resource")
			gogsCustomResource := gogscustomresource.NewGogsCustomResource(instance, "gogs-server", instance.Namespace)
			if err := r.client.Create(context.TODO(), gogsCustomResource); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			}
		}
	}

	//////////////////////////
	// CodeReady Workspaces
	//////////////////////////
	enabledWorkspaces := instance.Spec.Workspaces.Enabled
	enabledOpenShiftoAuth := instance.Spec.Workspaces.OpenShiftoAuth

	if enabledWorkspaces {
		workspacesNamespace := deployment.NewNamespace(instance, "workspaces")
		if err := r.client.Create(context.TODO(), workspacesNamespace); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Workspaces Projects")
		}

		workspacesCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "checlusters.org.eclipse.che", "org.eclipse.che", "CheCluster", "CheClusterList", "checlusters", "checluster", "v1", nil, nil)
		if err := r.client.Create(context.TODO(), workspacesCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Workspaces Custom Resource Definition")
		}

		workspacesServiceAccount := deployment.NewServiceAccount(instance, "workspaces-operator", workspacesNamespace.Name)
		if err := r.client.Create(context.TODO(), workspacesServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Workspaces Service Account")
		}

		workspacesClusterRole := deployment.NewClusterRole(instance, "workspaces-operator", workspacesNamespace.Name, deployment.WorkspacesRules())
		if err := r.client.Create(context.TODO(), workspacesClusterRole); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Workspaces Cluster Role")
		}

		workspacesClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "workspaces-operator", workspacesNamespace.Name, "workspaces-operator", "workspaces-operator", "ClusterRole")
		if err := r.client.Create(context.TODO(), workspacesClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Workspaces Cluster Role Binding")
		}

		commands := []string{
			"che-operator",
		}
		workspacesOperator := deployment.NewOperatorDeployment(instance, "workspaces-operator", workspacesNamespace.Name, "registry.redhat.io/codeready-workspaces/server-operator-rhel8:1.2", "workspaces-operator", 60000, commands, nil, nil, nil)
		if err := r.client.Create(context.TODO(), workspacesOperator); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Workspaces Operator")
		}

		workspacesCustomResource := checlustercustomresource.NewCheClusterCustomResource(instance, "codeready", workspacesNamespace.Name, "registry.redhat.io/codeready-workspaces/server-rhel8", "1.2", false, false, enabledOpenShiftoAuth)
		if err := r.client.Create(context.TODO(), workspacesCustomResource); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Workspaces Custom Resource")
		}

		openshiftStackImageStream := deployment.NewImageStream(instance, "che-cloud-native", "openshift", "quay.io/mcouliba/che-cloud-native", "ocp4")
		if err := r.client.Create(context.TODO(), openshiftStackImageStream); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			reqLogger.Info("Created Cloud Native Stack Image Stream (OCP4)")
		}

		var (
			body           []byte
			err            error
			url            string
			httpResponse   *http.Response
			httpRequest    *http.Request
			retries        = 60
			codereadyToken util.Token
			masterToken    util.Token
			client         = &http.Client{}
			stackResponse  = codereadystack.Stack{}
		)

		// Wait for CodeReady Workspaces to be running
		for retries > 0 {
			httpResponse, err = http.Get("http://codeready-workspaces." + appsHostnameSuffix + "/api/system/state")
			if err != nil {
				retries--
			} else {
				break
			}
			reqLogger.Info(fmt.Sprintf("Waiting for Workspaces to be up and running (%d retries left)", retries))
			time.Sleep(30 * time.Second)
		}

		if httpResponse == nil {
			return reconcile.Result{}, err
		}

		url = "http://keycloak-workspaces." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"
		httpRequest, err = http.NewRequest("POST", url, strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			return reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusOK {
			if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
				return reconcile.Result{}, err
			}
		}

		if !enabledOpenShiftoAuth {
			openshiftUserPassword := instance.Spec.Guide.OpenshiftUserPassword
			for id := 1; id <= users; id++ {
				username := fmt.Sprintf("user%d", id)
				body, err = json.Marshal(codereadyuser.NewCodeReadyUser(instance, username, openshiftUserPassword))
				if err != nil {
					return reconcile.Result{}, err
				}

				httpRequest, err = http.NewRequest("POST", "http://keycloak-workspaces."+appsHostnameSuffix+"/auth/admin/realms/codeready/users", bytes.NewBuffer(body))
				httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)
				httpRequest.Header.Set("Content-Type", "application/json")

				httpResponse, err = client.Do(httpRequest)
				if err != nil {
					reqLogger.Info("Error when creating " + username + " for CodeReady Workspaces")
					return reconcile.Result{}, err
				}
				if httpResponse.StatusCode == http.StatusCreated {
					reqLogger.Info("Created " + username + " for CodeReady Workspaces")
				}
			}
		}

		httpRequest, err = http.NewRequest("POST", "http://keycloak-workspaces."+appsHostnameSuffix+"/auth/realms/codeready/protocol/openid-connect/token", strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			reqLogger.Info("Error when getting Workspaces Access Token")
			return reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusOK {
			if err := json.NewDecoder(httpResponse.Body).Decode(&codereadyToken); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Workspaces Factory
		body, err = json.Marshal(codereadyfactory.NewDebuggingFactory(openshiftConsoleURL, openshiftAPIURL, appsHostnameSuffix, instance.Spec.Guide.OpenshiftUserPassword))
		if err != nil {
			return reconcile.Result{}, err
		}

		httpRequest, err = http.NewRequest("POST", "http://codeready-workspaces."+appsHostnameSuffix+"/api/factory", bytes.NewBuffer(body))
		httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
		httpRequest.Header.Set("Content-Type", "application/json")

		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			return reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusCreated || httpResponse.StatusCode == http.StatusOK {
			reqLogger.Info("Created Debugging Factory")
		}

		body, err = json.Marshal(codereadystack.NewCloudNativeStack(instance))
		if err != nil {
			return reconcile.Result{}, err
		}

		httpRequest, err = http.NewRequest("POST", "http://codeready-workspaces."+appsHostnameSuffix+"/api/stack", bytes.NewBuffer(body))
		httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
		httpRequest.Header.Set("Content-Type", "application/json")

		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			return reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusCreated || httpResponse.StatusCode == http.StatusOK {
			reqLogger.Info("Created Cloud Native Stack")

			if err := json.NewDecoder(httpResponse.Body).Decode(&stackResponse); err != nil {
				return reconcile.Result{}, err
			}

			body, err = json.Marshal(codereadystack.NewCloudNativeStackPermission(instance, stackResponse.ID))
			if err != nil {
				return reconcile.Result{}, err
			}

			httpRequest, err = http.NewRequest("POST", "http://codeready-workspaces."+appsHostnameSuffix+"/api/permissions", bytes.NewBuffer(body))
			httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
			httpRequest.Header.Set("Content-Type", "application/json")

			httpResponse, err = client.Do(httpRequest)
			if err != nil {
				return reconcile.Result{}, err
			}
			if httpResponse.StatusCode == http.StatusCreated {
				reqLogger.Info("Granted Cloud Native Stack")
			}

		}

	}

	//Success
	return reconcile.Result{}, nil
}
