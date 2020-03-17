package workshop

import (
	"context"
	"regexp"

	argocd "github.com/argoproj-labs/argocd-operator/pkg/apis/argoproj/v1alpha1"
	che "github.com/eclipse/che-operator/pkg/apis/org/v1"
	imagev1 "github.com/openshift/api/image/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	olmv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	ompv1 "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	gogscustomresource "github.com/redhat/openshift-workshop-operator/pkg/customresource/gogs"
	smcp "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshcontrolplane"
	smmr "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshmemberroll"
	nexus "github.com/redhat/openshift-workshop-operator/pkg/deployment/nexus"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
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

var (
	log = logf.Log.WithName("controller_workshop")
)

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

	// register CatalogSourceConfig in the scheme
	if err := ompv1.SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// register OperatorGroup in the scheme
	if err := olmv1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// register Subscription in the scheme
	if err := olmv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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

	// register OpenShift OAuthClient in the scheme
	if err := oauthv1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	// register Custom Resource Definition and Custom Resource in the scheme
	if err := apiextensionsv1beta1.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := argocd.SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := nexus.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := gogscustomresource.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	if err := che.SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
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

	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	}); err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbac.Role{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbac.ClusterRole{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbac.ClusterRoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbac.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openshiftv1alpha1.Workshop{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
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

	//////////////////////////
	// Variables
	//////////////////////////
	var (
		openshiftConsoleURL string
		openshiftAPIURL     string
		appsHostnameSuffix  string
	)
	// extract app route suffix from openshift-console
	openshiftConsoleRouteFound := &routev1.Route{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "console", Namespace: "openshift-console"}, openshiftConsoleRouteFound); err != nil {
		logrus.Errorf("Failed to get OpenShift Console: %s", err)
		return reconcile.Result{}, err
	}
	openshiftConsoleURL = "https://" + openshiftConsoleRouteFound.Spec.Host

	re := regexp.MustCompile("^console-openshift-console\\.apps\\.(.*?)$")
	match := re.FindStringSubmatch(openshiftConsoleRouteFound.Spec.Host)
	openshiftAPIURL = "https://api." + match[1]

	re = regexp.MustCompile("^console-openshift-console\\.(.*?)$")
	match = re.FindStringSubmatch(openshiftConsoleRouteFound.Spec.Host)
	appsHostnameSuffix = match[1]

	users := instance.Spec.User.Number
	if users < 0 {
		users = 0
	}

	//////////////////////////
	// Projects
	//////////////////////////
	if result, err := r.reconcileProject(instance, users); err != nil {
		return result, err
	}

	//////////////////////////
	// Workshopper
	//////////////////////////
	if result, err := r.reconcileWorkshopper(instance, users, appsHostnameSuffix,
		openshiftConsoleURL, openshiftAPIURL); err != nil {
		return result, err
	}

	//////////////////////////
	// Bookbag
	//////////////////////////
	if result, err := r.reconcileBookbag(instance, users, appsHostnameSuffix,
		openshiftConsoleURL); err != nil {
		return result, err
	}

	//////////////////////////
	// Etherpad
	//////////////////////////
	if result, err := r.reconcileEtherpad(instance, users, appsHostnameSuffix); err != nil {
		return result, err
	}

	//////////////////////////
	// Nexus
	//////////////////////////
	if result, err := r.reconcileNexus(instance); err != nil {
		return result, err
	}

	//////////////////////////
	// Pipeline
	//////////////////////////
	if result, err := r.reconcilePipeline(instance); err != nil {
		return result, err
	}

	//////////////////////////
	// Gogs
	//////////////////////////
	if result, err := r.reconcileGogs(instance); err != nil {
		return result, err
	}

	//////////////////////////
	// Argo CD
	//////////////////////////
	if result, err := r.reconcileArgoCD(instance, users, appsHostnameSuffix,
		openshiftConsoleURL, openshiftAPIURL); err != nil {
		return result, err
	}

	//////////////////////////
	// CodeReadyWorkspace
	//////////////////////////
	if result, err := r.reconcileCodeReadyWorkspace(instance, users, appsHostnameSuffix,
		openshiftConsoleURL, openshiftAPIURL); err != nil {
		return result, err
	}

	//////////////////////////
	// Service Mesh
	//////////////////////////
	if result, err := r.reconcileServiceMesh(instance, users); err != nil {
		return result, err
	}

	//////////////////////////
	// Serverless
	//////////////////////////
	if result, err := r.reconcileServerless(instance); err != nil {
		return result, err
	}

	//////////////////////////
	// Istio Workspace
	//////////////////////////
	if result, err := r.reconcileIstioWorkspace(instance, users); err != nil {
		return result, err
	}

	//Success
	reqLogger.Info("Workshop Reconciled")
	return reconcile.Result{}, nil
}
