package workshop

import (
	"context"
	"encoding/json"
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

	cheSubscription := deployment.NewSubscription(instance, "eclipse-che", cheNamespace.Name, "eclipse-che.v7.1.0")
	if err := r.client.Create(context.TODO(), cheSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", cheSubscription.Name)
	}

	// FIX - Workspaces fail to start with certain configurations of StorageClass
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

	var (
		// 	body         []byte
		// 	err          error
		// url          string
		// httpResponse *http.Response
		// httpRequest  *http.Request
		// 	retries      = 60
		// 	// codereadyToken util.Token
		// keycloakMasterTokenURL   = "http://keycloak-" + cheNamespace.Name + "." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"
		// keycloakTokenExchangeURL = "http://keycloak-" + cheNamespace.Name + "." + appsHostnameSuffix + "/auth/admin/realms/che/identity-provider/instances/openshift-v4/management/permissions"
		// oauthOpenShiftURL        = "https://oauth-openshift." + appsHostnameSuffix + "/oauth/authorize?client_id=openshift-challenging-client&response_type=token"

		// masterToken util.Token
		// client      = &http.Client{
		// 	Transport: &http.Transport{
		// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// 	}}
		// 	// stackResponse  = codereadystack.Stack{}
		timeout = 120
	)

	// FIX - Enabled Keycloak Token Exchange
	operatorImageFixed := "quay.io/dfestal/che-operator:enable-token-exchange"
	time.Sleep(time.Duration(1) * time.Second)
	cheCSV, err := r.GetEffectiveCSV(instance, "eclipse-che.v7.1.0", cheNamespace.Name)
	if err != nil {
		logrus.Errorf("Failed to get ClusterServiceVersion : %s", "eclipse-che.v7.1.0")
		logrus.Infof("Waiting for %s ClusterServiceVersion to be created (%v seconds)", "eclipse-che.v7.1.0", timeout)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * time.Duration(timeout)}, err
	}

	var strategySpecJSON map[string][]interface{}
	json.Unmarshal(cheCSV.Spec.InstallStrategy.StrategySpecRaw, &strategySpecJSON)

	if cheCSV.ObjectMeta.Annotations["containerImage"] != operatorImageFixed ||
		strategySpecJSON["deployments"][0].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["image"] != operatorImageFixed {
		cheCSV.ObjectMeta.Annotations["containerImage"] = operatorImageFixed
		strategySpecJSON["deployments"][0].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["image"] = operatorImageFixed
		cheCSV.Spec.InstallStrategy.StrategySpecRaw, _ = json.Marshal(strategySpecJSON)

		if err := r.client.Update(context.TODO(), cheCSV); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Updated '%s' ClusterServiceVersion (Fix)", "eclipse-che.v7.1.0")
		}
	}

	// Wait for Che Operator to be running
	time.Sleep(time.Duration(1) * time.Second)
	cheOperatorDeployment, err := r.GetEffectiveDeployment(instance, "che-operator", cheNamespace.Name)
	if err != nil {
		logrus.Infof("Waiting for OLM to create che-operator deployment (%v seconds)", timeout)
		time.Sleep(time.Duration(timeout) * time.Second)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * time.Duration(timeout)}, err
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

	// // Get Keycloak Admin Token
	// httpRequest, err = http.NewRequest("POST", keycloakMasterTokenURL, strings.NewReader("username=admin&password=gstd9f9oDDhN&grant_type=password&client_id=admin-cli"))
	// httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// httpResponse, err = client.Do(httpRequest)
	// if err != nil {
	// 	logrus.Errorf("Error to get the master token from che keycloak (%v)", err)
	// 	return reconcile.Result{}, err
	// }
	// defer httpResponse.Body.Close()
	// if httpResponse.StatusCode == http.StatusOK {
	// 	if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
	// 		logrus.Errorf("Error to get the master token from che keycloak (%v)", err)
	// 		return reconcile.Result{}, err
	// 	}
	// 	logrus.Infof("Got Keycloak Master Token")
	// } else {
	// 	logrus.Errorf("Error to get the master token from che keycloak (%d)", httpResponse.StatusCode)
	// 	return reconcile.Result{}, err
	// }

	// // Enable Token Exchange
	// httpRequest, err = http.NewRequest("PUT", keycloakTokenExchangeURL, bytes.NewBuffer([]byte("{\"enabled\" : true}")))
	// httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)
	// httpRequest.Header.Set("Content-Type", "application/json")

	// httpResponse, err = client.Do(httpRequest)
	// if err != nil {
	// 	logrus.Errorf("Error when enabling Keycloak Token Exchange (%v)", err)
	// 	return reconcile.Result{}, err
	// }
	// if httpResponse.StatusCode == http.StatusOK {
	// 	logrus.Infof("Enabled Keycloak Token Exchange")
	// } else {
	// 	logrus.Errorf("Error when enabling Keycloak Token Exchange (%d)", httpResponse.StatusCode)
	// 	return reconcile.Result{}, err
	// }

	// TODO - Add Policy and link to che-public

	// TODO - TOKEN
	// openshiftUserPassword := instance.Spec.UserPassword
	// for id := 1; id <= users; id++ {
	// 	username := fmt.Sprintf("user%d", id)

	// 	httpRequest, err = http.NewRequest("GET", oauthOpenShiftURL, nil)
	// 	httpRequest.Header.Set("Authorization", "Basic "+util.GetBasicAuth(username, openshiftUserPassword))
	// 	httpRequest.Header.Set("X-CSRF-Token", "xxx")

	// 	httpResponse, err = client.Do(httpRequest)
	// 	if err != nil {
	// 		logrus.Errorf("Error when getting Token Exchange for %s: %v", username, err)
	// 		return reconcile.Result{}, err
	// 	}
	// 	if httpResponse.StatusCode == http.StatusOK {
	// 		bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
	// 		if err != nil {
	// 			logrus.Errorf("BODY ERROR")
	// 		}
	// 		logrus.Infof("Got Token Exchange for %s: %s", username, string(bodyBytes))
	// 	} else {
	// 		logrus.Errorf("Error when getting Token Exchange for %s (%d)", username, httpResponse.StatusCode)
	// 		return reconcile.Result{}, err
	// 	}
	// }

	//Success
	return reconcile.Result{}, nil
}
