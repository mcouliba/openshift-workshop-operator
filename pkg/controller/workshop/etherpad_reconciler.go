package workshop

import (
	"context"
	"fmt"
	"strings"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Reconciling Etherpad
func (r *ReconcileWorkshop) reconcileEtherpad(instance *openshiftv1alpha1.Workshop, users int, appsHostnameSuffix string) error {
	enabledEtherpad := instance.Spec.Infrastructure.Etherpad.Enabled

	if enabledEtherpad {
		if err := r.addEtherpad(instance, users, appsHostnameSuffix); err != nil {
			return err
		}
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addEtherpad(instance *openshiftv1alpha1.Workshop, users int, appsHostnameSuffix string) error {
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
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad SQL Secret")
	}

	etherpadDatabasePersistentVolumeClaim := deployment.NewPersistentVolumeClaim(instance, "etherpad-mysql", instance.Namespace, "512Mi")
	if err := r.client.Create(context.TODO(), etherpadDatabasePersistentVolumeClaim); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad SQL Persistent Volume Claim")
	}

	etherpadDatabaseDeployment := deployment.NewEtherpadDatabaseDeployment(instance, "etherpad-mysql", instance.Namespace)
	if err := r.client.Create(context.TODO(), etherpadDatabaseDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad SQL Database")
	}

	etherpadDatabaseService := deployment.NewService(instance, "etherpad-mysql", instance.Namespace, []string{"mysql"}, []int32{3306})
	if err := r.client.Create(context.TODO(), etherpadDatabaseService); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad SQL Service")
	}

	settings := map[string]string{
		"settings.json": deployment.NewEtherpadSettingsJson(instance, userEndpointStr.String()),
	}
	etherpadConfigMap := deployment.NewConfigMap(instance, "etherpad-settings", instance.Namespace, settings)
	if err := r.client.Create(context.TODO(), etherpadConfigMap); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad ConfigMap")
	}

	etherpadDeployment := deployment.NewEtherpadDeployment(instance, "etherpad", instance.Namespace)
	if err := r.client.Create(context.TODO(), etherpadDeployment); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad Deployment")
	}

	etherpadService := deployment.NewService(instance, "etherpad", instance.Namespace, []string{"http"}, []int32{9001})
	if err := r.client.Create(context.TODO(), etherpadService); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad Service")
	}

	etherpadRoute := deployment.NewRoute(instance, "etherpad", instance.Namespace, "etherpad", 9001)
	if err := r.client.Create(context.TODO(), etherpadRoute); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Etherpad Route")
	}

	//Success
	return nil
}
