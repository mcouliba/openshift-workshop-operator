package workshop

import (
	"context"
	"fmt"
	"strings"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Etherpad
func (r *ReconcileWorkshop) reconcileEtherpad(instance *openshiftv1alpha1.Workshop, users int, appsHostnameSuffix string) (reconcile.Result, error) {
	enabledEtherpad := instance.Spec.Infrastructure.Etherpad.Enabled

	if enabledEtherpad {
		if result, err := r.addEtherpad(instance, users, appsHostnameSuffix); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.Etherpad != util.OperatorStatus.Installed {
			instance.Status.Etherpad = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addEtherpad(instance *openshiftv1alpha1.Workshop, users int, appsHostnameSuffix string) (reconcile.Result, error) {
	reqLogger := log.WithName("Etherpad")

	var userEndpointStr strings.Builder
	for id := 1; id <= users; id++ {
		userEndpointStr.WriteString(fmt.Sprintf("You are user%d\t|\thttp://guide-infra%d.%s\t|\t<INSERT_YOUR_NAME>\n", id, id, appsHostnameSuffix))
	}

	databaseCredentials := map[string]string{
		"database-name":          "sampledb",
		"database-password":      "admin",
		"database-root-password": "admin",
		"database-user":          "admin",
	}
	etherpadDatabaseSecret := deployment.NewSecretStringData(instance, "etherpad-mysql", instance.Namespace, databaseCredentials)
	if err := r.client.Create(context.TODO(), etherpadDatabaseSecret); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Secret", etherpadDatabaseSecret.Name)
	}

	etherpadDatabasePersistentVolumeClaim := deployment.NewPersistentVolumeClaim(instance, "etherpad-mysql", instance.Namespace, "512Mi")
	if err := r.client.Create(context.TODO(), etherpadDatabasePersistentVolumeClaim); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Persistent Volume Claim", etherpadDatabasePersistentVolumeClaim.Name)
	}

	etherpadDatabaseDeployment := deployment.NewEtherpadDatabaseDeployment(instance, "etherpad-mysql", instance.Namespace)
	if err := r.client.Create(context.TODO(), etherpadDatabaseDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Database", etherpadDatabaseDeployment.Name)
	}

	etherpadDatabaseService := deployment.NewService(instance, "etherpad-mysql", instance.Namespace, []string{"mysql"}, []int32{3306})
	if err := r.client.Create(context.TODO(), etherpadDatabaseService); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Etherpad SQL Service")
		logrus.Infof("Created %s Service", etherpadDatabaseService.Name)
	}

	settings := map[string]string{
		"settings.json": deployment.NewEtherpadSettingsJson(instance, userEndpointStr.String()),
	}

	labels := map[string]string{
		"app.kubernetes.io/name":    "etherpad-settings",
		"app.kubernetes.io/part-of": "etherpad",
	}

	etherpadConfigMap := deployment.NewConfigMap(instance, "etherpad-settings", instance.Namespace, labels, settings)
	if err := r.client.Create(context.TODO(), etherpadConfigMap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s ConfigMap", etherpadConfigMap.Name)
	}

	etherpadDeployment := deployment.NewEtherpadDeployment(instance, "etherpad", instance.Namespace)
	if err := r.client.Create(context.TODO(), etherpadDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Deployment", etherpadDeployment.Name)
	}

	etherpadService := deployment.NewService(instance, "etherpad", instance.Namespace, []string{"http"}, []int32{9001})
	if err := r.client.Create(context.TODO(), etherpadService); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service", etherpadService.Name)
	}

	etherpadRoute := deployment.NewRoute(instance, "etherpad", instance.Namespace, "etherpad", 9001)
	if err := r.client.Create(context.TODO(), etherpadRoute); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Route", etherpadRoute.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
