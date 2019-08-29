package workshop

import (
	"context"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	che "github.com/redhat/openshift-workshop-operator/pkg/deployment/che"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Che
func (r *ReconcileWorkshop) reconcileChe(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	enabledChe := instance.Spec.Che.Enabled

	if enabledChe {
		if result, err := r.addChe(instance, users, appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addChe(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	var (
		// 	body         []byte
		// 	err          error
		// 	url          string
		// 	httpResponse *http.Response
		// 	httpRequest  *http.Request
		// 	retries      = 60
		// 	// codereadyToken util.Token
		// 	masterToken util.Token
		// 	client      = &http.Client{}
		// 	// stackResponse  = codereadystack.Stack{}
		reqLogger = log.WithName("Che")
		timeout   = 60
	)

	cheNamespace := deployment.NewNamespace(instance, "eclipse-che")
	if err := r.client.Create(context.TODO(), cheNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", cheNamespace.Name)
	}

	cheCatalogSourceConfig := deployment.NewCatalogSourceConfig(instance, "eclipse-che", cheNamespace.Name)
	if err := r.client.Create(context.TODO(), cheCatalogSourceConfig); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s CatalogSourceConfig", cheCatalogSourceConfig.Name)
	}

	cheOperatorGroup := deployment.NewOperatorGroup(instance, "eclipse-che", cheNamespace.Name)
	if err := r.client.Create(context.TODO(), cheOperatorGroup); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OperatorGroup", cheOperatorGroup.Name)
	}

	cheSubscription := deployment.NewSubscription(instance, "eclipse-che", cheNamespace.Name, "eclipse-che.v7.0.0")
	if err := r.client.Create(context.TODO(), cheSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", cheSubscription.Name)
	}

	// Fixed Workspaces fail to start with certain configurations of StorageClass
	configMapData := map[string]string{
		"CHE_INFRA_KUBERNETES_PVC_WAIT__BOUND":                  "false",
		"CHE_INFRA_KUBERNETES_SERVICE__ACCOUNT__NAME":           "che-workspace",
		"CHE_INFRA_KUBERNETES_WORKSPACE__UNRECOVERABLE__EVENTS": "FailedMount,FailedScheduling,MountVolume.SetUp failed,Failed to pull image",
		"CHE_PREDEFINED_STACKS_RELOAD__ON__START":               "true",
		"CHE_WORKSPACE_AGENT_DEV_INACTIVE__STOP__TIMEOUT__MS":   "-1",
		"CHE_WORKSPACE_AUTO_START":                              "true",
	}
	workspacesCustomConfigMap := deployment.NewConfigMap(instance, "custom", cheNamespace.Name, configMapData)
	if err := r.client.Create(context.TODO(), workspacesCustomConfigMap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created Che Custom ConfigMap")
	}

	// Wait for Che Operator to be running
	cheOperatorDeployment, err := r.GetEffectiveDeployment(instance, "che-operator", cheNamespace.Name)
	if err != nil {
		logrus.Errorf("Failed to get che-operator deployment: %v", err)
		logrus.Infof("Waiting for OLM to create che-operator deployment (%v seconds)", timeout)
		time.Sleep(time.Duration(timeout) * time.Second)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, err
	}

	if cheOperatorDeployment.Status.AvailableReplicas != 1 {
		scaled := k8sclient.GetDeploymentStatus("che-operator", cheNamespace.Name)
		if !scaled {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, err
		}
	}

	cheCustomResource := che.NewCustomResource(instance, "eclipse-che", cheNamespace.Name)
	if err := r.client.Create(context.TODO(), cheCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", cheCustomResource.Name)
	}

	// Wait for CodeReady Workspaces to be running
	workspacesDeployment, err := r.GetEffectiveDeployment(instance, "che", cheNamespace.Name)
	if err != nil {
		reqLogger.Error(err, "Failed to get codeready deployment")
		logrus.Infof("Waiting for Che Operator to build resources (%v seconds)", timeout)
		time.Sleep(time.Duration(timeout) * time.Second)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, err
	}

	if workspacesDeployment.Status.AvailableReplicas != 1 {
		scaled := k8sclient.GetDeploymentStatus("che", cheNamespace.Name)
		if !scaled {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, err
		}
	}

	// openshiftStackImageStream := deployment.NewImageStream(instance, "che-cloud-native", "openshift", "quay.io/mcouliba/che-cloud-native:ocp4", "ocp4")
	// if err := r.client.Create(context.TODO(), openshiftStackImageStream); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	reqLogger.Info("Created Cloud Native Stack Image Stream (OCP4)")
	// }

	// url = "http://keycloak-che." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"
	// httpRequest, err = http.NewRequest("POST", url, strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
	// httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// httpResponse, err = client.Do(httpRequest)
	// if err != nil {
	// 	return reconcile.Result{}, err
	// }
	// defer httpResponse.Body.Close()
	// if httpResponse.StatusCode == http.StatusOK {
	// 	if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
	// 		return reconcile.Result{}, err
	// 	}
	// }

	// if !instance.Spec.Che.OpenShiftoAuth {
	// 	openshiftUserPassword := instance.Spec.UserPassword
	// 	for id := 1; id <= users; id++ {
	// 		username := fmt.Sprintf("user%d", id)
	// 		body, err = json.Marshal(codereadyuser.NewCodeReadyUser(instance, username, openshiftUserPassword))
	// 		if err != nil {
	// 			return reconcile.Result{}, err
	// 		}

	// 		httpRequest, err = http.NewRequest("POST", "http://keycloak-che."+appsHostnameSuffix+"/auth/admin/realms/codeready/users", bytes.NewBuffer(body))
	// 		httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)
	// 		httpRequest.Header.Set("Content-Type", "application/json")

	// 		httpResponse, err = client.Do(httpRequest)
	// 		if err != nil {
	// 			reqLogger.Info("Error when creating " + username + " for CodeReady Che")
	// 			return reconcile.Result{}, err
	// 		}
	// 		if httpResponse.StatusCode == http.StatusCreated {
	// 			reqLogger.Info("Created " + username + " for CodeReady Che")
	// 		}
	// 	}
	// }

	// httpRequest, err = http.NewRequest("POST", "http://keycloak-che."+appsHostnameSuffix+"/auth/realms/codeready/protocol/openid-connect/token", strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
	// httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// httpResponse, err = client.Do(httpRequest)
	// if err != nil {
	// 	reqLogger.Info("Error when getting Che Access Token")
	// 	return reconcile.Result{}, err
	// }
	// defer httpResponse.Body.Close()
	// if httpResponse.StatusCode == http.StatusOK {
	// 	if err := json.NewDecoder(httpResponse.Body).Decode(&codereadyToken); err != nil {
	// 		return reconcile.Result{}, err
	// 	}
	// }

	// // Che Factory
	// body, err = json.Marshal(codereadyfactory.NewDebuggingFactory(openshiftConsoleURL, openshiftAPIURL, appsHostnameSuffix, instance.Spec.UserPassword))
	// if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// httpRequest, err = http.NewRequest("POST", "http://codeready-che."+appsHostnameSuffix+"/api/factory", bytes.NewBuffer(body))
	// httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
	// httpRequest.Header.Set("Content-Type", "application/json")

	// httpResponse, err = client.Do(httpRequest)
	// if err != nil {
	// 	return reconcile.Result{}, err
	// }
	// defer httpResponse.Body.Close()
	// if httpResponse.StatusCode == http.StatusCreated || httpResponse.StatusCode == http.StatusOK {
	// 	reqLogger.Info("Created Debugging Factory")
	// }

	// body, err = json.Marshal(codereadystack.NewCloudNativeStack(instance))
	// if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// httpRequest, err = http.NewRequest("POST", "http://codeready-che."+appsHostnameSuffix+"/api/stack", bytes.NewBuffer(body))
	// httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
	// httpRequest.Header.Set("Content-Type", "application/json")

	// httpResponse, err = client.Do(httpRequest)
	// if err != nil {
	// 	return reconcile.Result{}, err
	// }
	// defer httpResponse.Body.Close()
	// if httpResponse.StatusCode == http.StatusCreated || httpResponse.StatusCode == http.StatusOK {
	// 	reqLogger.Info("Created Cloud Native Stack")

	// 	if err := json.NewDecoder(httpResponse.Body).Decode(&stackResponse); err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	body, err = json.Marshal(codereadystack.NewCloudNativeStackPermission(instance, stackResponse.ID))
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	httpRequest, err = http.NewRequest("POST", "http://codeready-che."+appsHostnameSuffix+"/api/permissions", bytes.NewBuffer(body))
	// 	httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
	// 	httpRequest.Header.Set("Content-Type", "application/json")

	// 	httpResponse, err = client.Do(httpRequest)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}
	// 	if httpResponse.StatusCode == http.StatusCreated {
	// 		reqLogger.Info("Granted Cloud Native Stack")
	// 	}

	// }

	//Success
	return reconcile.Result{}, nil
}
