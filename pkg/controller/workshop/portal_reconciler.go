package workshop

import (
	"context"
	"reflect"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Portal
func (r *ReconcileWorkshop) reconcilePortal(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {

	if result, err := r.addRedis(instance); err != nil {
		return result, err
	}

	if result, err := r.addUpdateUsernameDistribution(instance, users, appsHostnameSuffix, openshiftConsoleURL); err != nil {
		return result, err
	}

	// Installed
	if instance.Status.UsernameDistribution != util.OperatorStatus.Installed {
		instance.Status.UsernameDistribution = util.OperatorStatus.Installed
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			logrus.Errorf("Failed to update Workshop status: %s", err)
			return reconcile.Result{}, nil
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addRedis(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {

	namespace := deployment.NewNamespace(instance, "workshop-guides")
	if err := r.client.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", namespace.Name)
	}

	serviceName := "redis"
	labels := map[string]string{
		"app":                       serviceName,
		"app.kubernetes.io/part-of": "portal",
	}

	credentials := map[string]string{
		"database-password": "redis",
	}
	secret := deployment.NewSecretStringData(instance, serviceName, namespace.Name, credentials)
	if err := r.client.Create(context.TODO(), secret); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Secret", secret.Name)
	}

	persistentVolumeClaim := deployment.NewPersistentVolumeClaim(instance, serviceName, namespace.Name, "512Mi")
	if err := r.client.Create(context.TODO(), persistentVolumeClaim); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Persistent Volume Claim", persistentVolumeClaim.Name)
	}

	// Deploy/Update UsernameDistribution
	ocpDeployment := deployment.NewRedisDeployment(instance, "redis", namespace.Name, labels)
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
	service := deployment.NewService(instance, serviceName, namespace.Name, labels, []string{"http"}, []int32{6379})
	if err := r.client.Create(context.TODO(), service); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service", service.Name)
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addUpdateUsernameDistribution(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {

	namespace := deployment.NewNamespace(instance, "workshop-guides")
	if err := r.client.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", namespace.Name)
	}

	serviceName := "portal"
	redisServiceName := "redis"
	labels := map[string]string{
		"app":                       serviceName,
		"app.kubernetes.io/part-of": "portal",
	}

	// Deploy/Update UsernameDistribution
	ocpDeployment := deployment.NewUsernameDistributionDeployment(instance, serviceName, namespace.Name, labels, redisServiceName, users, appsHostnameSuffix)
	if err := r.client.Create(context.TODO(), ocpDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Deployment", ocpDeployment.Name)
	} else if errors.IsAlreadyExists(err) {
		deploymentFound := &appsv1.Deployment{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ocpDeployment.Name, Namespace: namespace.Name}, deploymentFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(ocpDeployment.Spec.Template.Spec.Containers[0].Env, deploymentFound.Spec.Template.Spec.Containers[0].Env) ||
				!reflect.DeepEqual(ocpDeployment.Spec.Template.Spec.Containers[0].Image, deploymentFound.Spec.Template.Spec.Containers[0].Image) {
				// Update Guide
				if err := r.client.Update(context.TODO(), ocpDeployment); err != nil {
					return reconcile.Result{}, err
				}
				logrus.Infof("Updated %s Deployment", ocpDeployment.Name)
			}
		}
	}

	// Create Service
	service := deployment.NewService(instance, serviceName, namespace.Name, labels, []string{"http"}, []int32{8080})
	if err := r.client.Create(context.TODO(), service); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service", service.Name)
	}

	// Create Route
	route := deployment.NewRoute(instance, serviceName, namespace.Name, labels, serviceName, 8080)
	if err := r.client.Create(context.TODO(), route); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Route", route.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
