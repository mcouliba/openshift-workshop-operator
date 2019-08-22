package workshop

import (
	"context"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	k8sclient = GetK8Client()
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
	reqLogger := log.WithName("Che")

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

	// cheCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "checlusters.org.eclipse.che", "org.eclipse.che", "CheCluster", "CheClusterList", "checlusters", "checluster", "v1", nil, nil)
	// if err := r.client.Create(context.TODO(), cheCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	reqLogger.Info("Created Che Custom Resource Definition")
	// }

	// cheServiceAccount := deployment.NewServiceAccount(instance, "che-operator", cheNamespace.Name)
	// if err := r.client.Create(context.TODO(), cheServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	reqLogger.Info("Created Che Service Account")
	// }

	// cheClusterRole := deployment.NewClusterRole(instance, "che-operator", cheNamespace.Name, che.CheRules())
	// if err := r.client.Create(context.TODO(), cheClusterRole); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	reqLogger.Info("Created Che Cluster Role")
	// }

	// cheClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "che-operator", cheNamespace.Name, "che-operator", "che-operator", "ClusterRole")
	// if err := r.client.Create(context.TODO(), cheClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	reqLogger.Info("Created Che Cluster Role Binding")
	// }

	// commands := []string{
	// 	"/usr/local/bin/che-operator",
	// }
	// cheOperator := deployment.NewOperatorDeployment(instance, "che-operator", cheNamespace.Name, instance.Spec.Che.OperatorImage, "che-operator", 60000, commands, nil, nil, nil)
	// if err := r.client.Create(context.TODO(), cheOperator); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	reqLogger.Info("Created Che Operator")
	// }

	cheCustomResource := deployment.NewCheClusterCustomResource(instance, "eclipse-che", cheNamespace.Name)
	if err := r.client.Create(context.TODO(), cheCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", cheCustomResource.Name)
	}

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
		timeout = 180
	)

	// Wait for CodeReady Workspaces to be running
	logrus.Infof("Waiting for Che Operator to build resources (%v seconds)", timeout)
	time.Sleep(time.Duration(timeout) * time.Second)

	workspacesDeployment, err := r.GetEffectiveDeployment(instance, "che", cheNamespace.Name)
	if err != nil {
		reqLogger.Error(err, "Failed to get codeready deployment")
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
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
