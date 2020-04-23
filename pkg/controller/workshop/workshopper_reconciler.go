package workshop

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	routev1 "github.com/openshift/api/route/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Workshopper
func (r *ReconcileWorkshop) reconcileWorkshopper(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	enabledWorkshopper := instance.Spec.Infrastructure.Workshopper.Enabled

	id := 1
	for {
		if id <= users && enabledWorkshopper {
			// Guide
			if result, err := r.addUpdateWorkshopper(instance, strconv.Itoa(id),
				appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL); err != nil {
				return result, err
			}
		} else {

			infraProjectName := fmt.Sprintf("infra%d", id)

			infraNamespace := deployment.NewNamespace(instance, infraProjectName)
			infraNamespaceFound := &corev1.Namespace{}
			infraNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: infraNamespace.Name}, infraNamespaceFound)

			if infraNamespaceErr != nil && errors.IsNotFound(infraNamespaceErr) {
				break
			}

			if result, err := r.deleteWorkshopper(instance, infraProjectName); err != nil {
				return result, err
			}
		}

		id++
	}

	// Installed
	if instance.Status.Workshopper != util.OperatorStatus.Installed {
		instance.Status.Workshopper = util.OperatorStatus.Installed
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			logrus.Errorf("Failed to update Workshop status: %s", err)
			return reconcile.Result{}, nil
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addUpdateWorkshopper(instance *openshiftv1alpha1.Workshop, userID string,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {

	labels := map[string]string{
		"app":                       "guide",
		"app.kubernetes.io/part-of": "workshopper",
	}

	infraProjectName := fmt.Sprintf("infra%s", userID)
	workshopperNamespace := deployment.NewNamespace(instance, infraProjectName)
	if err := r.client.Create(context.TODO(), workshopperNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Namespace", workshopperNamespace.Name)
	}

	// Deploy/Update Guide
	guideDeployment := deployment.NewWorkshopperDeployment(instance, "guide", infraProjectName, labels, userID, appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL)
	if err := r.client.Create(context.TODO(), guideDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created Guide Deployment for %s", workshopperNamespace.Name)
	} else if errors.IsAlreadyExists(err) {
		guideDeploymentFound := &appsv1.Deployment{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "guide", Namespace: infraProjectName}, guideDeploymentFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(guideDeployment.Spec.Template.Spec.Containers[0].Env, guideDeploymentFound.Spec.Template.Spec.Containers[0].Env) {
				// Update Guide
				if err := r.client.Update(context.TODO(), guideDeployment); err != nil {
					return reconcile.Result{}, err
				}
				logrus.Infof("Updated Guide Deployment for %s", workshopperNamespace.Name)
			}
		}
	}

	// Create Service
	guideService := deployment.NewService(instance, "guide", infraProjectName, labels, []string{"http"}, []int32{8080})
	if err := r.client.Create(context.TODO(), guideService); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created Guide Service for %s", workshopperNamespace.Name)
	}

	// Create Route
	guideRoute := deployment.NewRoute(instance, "guide", infraProjectName, labels, "guide", 8080)
	if err := r.client.Create(context.TODO(), guideRoute); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created Guide Route for %s", workshopperNamespace.Name)
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) deleteWorkshopper(instance *openshiftv1alpha1.Workshop, infraProjectName string) (reconcile.Result, error) {

	guideRouteFound := &routev1.Route{}
	guideRouteErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "guide", Namespace: infraProjectName}, guideRouteFound)
	if guideRouteErr == nil {
		// Delete Route
		if err := r.client.Delete(context.TODO(), guideRouteFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted Guide Route for %s", infraProjectName)
	}
	guideServiceFound := &corev1.Service{}
	guideServiceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "guide", Namespace: infraProjectName}, guideServiceFound)
	if guideServiceErr == nil {
		// Delete Service
		if err := r.client.Delete(context.TODO(), guideServiceFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted Guide Service for %s", infraProjectName)
	}

	guideDeploymentFound := &appsv1.Deployment{}
	guideDeploymentErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "guide", Namespace: infraProjectName}, guideDeploymentFound)
	if guideDeploymentErr == nil {
		// Undeploy Guide
		if err := r.client.Delete(context.TODO(), guideDeploymentFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted Guide Deployment for %s", infraProjectName)
	}

	workshopperNamespaceFound := &appsv1.Deployment{}
	workshopperNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: infraProjectName}, workshopperNamespaceFound)
	if workshopperNamespaceErr == nil {
		// Delete Namespace Infra
		if err := r.client.Delete(context.TODO(), workshopperNamespaceFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted %s Project", infraProjectName)
	}

	//Success
	return reconcile.Result{}, nil
}
