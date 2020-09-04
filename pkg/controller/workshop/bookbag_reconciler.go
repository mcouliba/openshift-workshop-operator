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

// Reconciling Bookbag
func (r *ReconcileWorkshop) reconcileBookbag(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {
	enabled := instance.Spec.Infrastructure.Bookbag.Enabled

	guidesNamespace := "workshop-guides"

	id := 1
	for {
		if id <= users && enabled {
			// Bookback
			if result, err := r.addUpdateBookbag(instance, strconv.Itoa(id), guidesNamespace,
				appsHostnameSuffix, openshiftConsoleURL); err != nil {
				return result, err
			}
		} else {

			bookbagName := fmt.Sprintf("bookbag-%d", id)

			ocpDeploymentFound := &appsv1.Deployment{}
			ocpDeploymentErr := r.client.Get(context.TODO(), types.NamespacedName{Name: bookbagName, Namespace: guidesNamespace}, ocpDeploymentFound)

			if ocpDeploymentErr != nil && errors.IsNotFound(ocpDeploymentErr) {
				break
			}

			if result, err := r.deleteBookbag(instance, strconv.Itoa(id), guidesNamespace); err != nil {
				return result, err
			}
		}

		id++
	}

	// Installed
	if instance.Status.Bookbag != util.OperatorStatus.Installed {
		instance.Status.Bookbag = util.OperatorStatus.Installed
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			logrus.Errorf("Failed to update Workshop status: %s", err)
			return reconcile.Result{}, nil
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addUpdateBookbag(instance *openshiftv1alpha1.Workshop, userID string,
	guidesNamespace string, appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {

	namespace := deployment.NewNamespace(instance, guidesNamespace)
	if err := r.client.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", namespace.Name)
	}

	bookbagName := fmt.Sprintf("user%s-bookbag", userID)
	labels := map[string]string{
		"app":                       bookbagName,
		"app.kubernetes.io/part-of": "bookbag",
	}

	// Create ConfigMap
	data := map[string]string{
		"gateway.sh":  "",
		"terminal.sh": "",
		"workshop.sh": "",
	}

	envConfigMap := deployment.NewConfigMap(instance, bookbagName+"-env", namespace.Name, labels, data)
	if err := r.client.Create(context.TODO(), envConfigMap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s ConfigMap", envConfigMap.Name)
	}

	varConfigMap := deployment.NewConfigMap(instance, bookbagName+"-vars", namespace.Name, labels, nil)
	if err := r.client.Create(context.TODO(), varConfigMap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s ConfigMap", varConfigMap.Name)
	}

	// Create Service Account
	serviceAccount := deployment.NewServiceAccount(instance, bookbagName, namespace.Name)
	if err := r.client.Create(context.TODO(), serviceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service Account", serviceAccount.Name)
	}

	// Create Role Binding
	roleBinding := deployment.NewRoleBindingSA(deployment.NewRoleBindingSAParameters{
		Name:               bookbagName,
		Namespace:          namespace.Name,
		ServiceAccountName: serviceAccount.Name,
		RoleName:           "admin",
		RoleKind:           "Role",
	})

	if err := r.client.Create(context.TODO(), roleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", roleBinding.Name)
	}

	// Deploy/Update Bookbag
	ocpDeployment := deployment.NewBookbagDeployment(instance, bookbagName, namespace.Name, labels, userID, appsHostnameSuffix, openshiftConsoleURL)
	if err := r.client.Create(context.TODO(), ocpDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Deployment", ocpDeployment.Name)
	} else if errors.IsAlreadyExists(err) {
		deploymentFound := &appsv1.Deployment{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ocpDeployment.Name, Namespace: namespace.Name}, deploymentFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(ocpDeployment.Spec.Template.Spec.Containers[0].Env, deploymentFound.Spec.Template.Spec.Containers[0].Env) {
				// Update Guide
				if err := r.client.Update(context.TODO(), ocpDeployment); err != nil {
					return reconcile.Result{}, err
				}
				logrus.Infof("Updated %s Deployment", ocpDeployment.Name)
			}
		}
	}

	// Create Service
	service := deployment.NewService(instance, bookbagName, namespace.Name, labels, []string{"http"}, []int32{10080})
	if err := r.client.Create(context.TODO(), service); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service", service.Name)
	}

	// Create Route
	route := deployment.NewRoute(instance, bookbagName, namespace.Name, labels, bookbagName, 10080)
	if err := r.client.Create(context.TODO(), route); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Route", route.Name)
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) deleteBookbag(instance *openshiftv1alpha1.Workshop, userID string, guidesNamespace string) (reconcile.Result, error) {

	bookbagName := fmt.Sprintf("user%s-bookbag", userID)

	routeFound := &routev1.Route{}
	routeErr := r.client.Get(context.TODO(), types.NamespacedName{Name: bookbagName, Namespace: guidesNamespace}, routeFound)
	if routeErr == nil {
		// Delete Route
		if err := r.client.Delete(context.TODO(), routeFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted %s Route", routeFound.Name)
	}

	serviceFound := &corev1.Service{}
	serviceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: bookbagName, Namespace: guidesNamespace}, serviceFound)
	if serviceErr == nil {
		// Delete Service
		if err := r.client.Delete(context.TODO(), serviceFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted %s Service", serviceFound.Name)
	}

	ocpDeploymentFound := &appsv1.Deployment{}
	ocpDeploymentErr := r.client.Get(context.TODO(), types.NamespacedName{Name: bookbagName, Namespace: guidesNamespace}, ocpDeploymentFound)
	if ocpDeploymentErr == nil {
		// Undeploy Guide
		if err := r.client.Delete(context.TODO(), ocpDeploymentFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted %s Deployment", ocpDeploymentFound.Name)
	}

	serviceAccountFound := &corev1.ServiceAccount{}
	serviceAccountErr := r.client.Get(context.TODO(), types.NamespacedName{Name: bookbagName, Namespace: guidesNamespace}, serviceAccountFound)
	if serviceAccountErr == nil {
		// Delete Service
		if err := r.client.Delete(context.TODO(), serviceAccountFound); err != nil {
			return reconcile.Result{}, err
		}
		logrus.Infof("Deleted %s Service Account", serviceAccountFound.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
