package workshop

import (
	"context"
	"io/ioutil"
	"strings"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// Reconciling Etherpad
func (r *ReconcileWorkshop) reconcileEtherpad(instance *openshiftv1alpha1.Workshop, userEndpointStr strings.Builder) error {
	var err error
	enabledEtherpad := instance.Spec.Etherpad.Enabled

	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "etherpad", Namespace: instance.Namespace}, found)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if enabledEtherpad {
		if err != nil && errors.IsNotFound(err) {
			if err = r.addEtherpad(instance, userEndpointStr); err != nil {
				return err
			}
		}
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addEtherpad(instance *openshiftv1alpha1.Workshop, userEndpointStr strings.Builder) error {
	reqLogger := log.WithName("Etherpad")

	reqLogger.Info("Creating Etherpad SQL Secret")
	databaseCredentials := map[string]string{
		"database-name":          "sampledb",
		"database-password":      "admin",
		"database-root-password": "admin",
		"database-user":          "admin",
	}
	etherpadDatabaseSecret := deployment.NewSecretStringData(instance, "etherpad-mysql", instance.Namespace, databaseCredentials)
	if err := r.client.Create(context.TODO(), etherpadDatabaseSecret); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	reqLogger.Info("Creating Etherpad SQL Persistent Volume Claim")
	etherpadDatabasePersistentVolumeClaim := deployment.NewPersistentVolumeClaim(instance, "etherpad-mysql", instance.Namespace, "512Mi")
	if err := r.client.Create(context.TODO(), etherpadDatabasePersistentVolumeClaim); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	reqLogger.Info("Creating Etherpad SQL Database")
	etherpadDatabaseDeployment := deployment.NewEtherpadDatabaseDeployment(instance, "etherpad-mysql", instance.Namespace)
	if err := r.client.Create(context.TODO(), etherpadDatabaseDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	reqLogger.Info("Creating Etherpad SQL Service")
	etherpadDatabaseService := deployment.NewService(instance, "etherpad-mysql", instance.Namespace, []string{"mysql"}, []int32{3306})
	if err := r.client.Create(context.TODO(), etherpadDatabaseService); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	reqLogger.Info("Creating Etherpad ConfigMap")
	defaultPadBytes, err := ioutil.ReadFile("template/etherpad.txt")
	if err != nil {
		return err
	}
	defaultPadText := strings.Replace(string(defaultPadBytes), "<USER_ENDPOINTS>", userEndpointStr.String(), 1)
	settings := map[string]string{
		"settings.json": deployment.NewEtherpadSettingsJson(instance, defaultPadText),
	}
	etherpadConfigMap := deployment.NewConfigMap(instance, "etherpad-settings", instance.Namespace, settings)
	if err := r.client.Create(context.TODO(), etherpadConfigMap); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	reqLogger.Info("Creating Etherpad Deployment")
	etherpadDeployment := deployment.NewEtherpadDeployment(instance, "etherpad", instance.Namespace)
	if err := r.client.Create(context.TODO(), etherpadDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	reqLogger.Info("Creating Etherpad Service")
	etherpadService := deployment.NewService(instance, "etherpad", instance.Namespace, []string{"http"}, []int32{9001})
	if err := r.client.Create(context.TODO(), etherpadService); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	reqLogger.Info("Creating Etherpad Route")
	etherpadRoute := deployment.NewRoute(instance, "etherpad", instance.Namespace, "etherpad", 9001)
	if err := r.client.Create(context.TODO(), etherpadRoute); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	//Success
	return nil
}
