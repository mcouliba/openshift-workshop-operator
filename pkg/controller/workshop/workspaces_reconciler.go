package workshop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	codereadyfactory "github.com/redhat/openshift-workshop-operator/pkg/codeready/factory"
	codereadystack "github.com/redhat/openshift-workshop-operator/pkg/codeready/stack"
	codereadyuser "github.com/redhat/openshift-workshop-operator/pkg/codeready/user"
	checlustercustomresource "github.com/redhat/openshift-workshop-operator/pkg/customresource/checluster"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Workspaces
func (r *ReconcileWorkshop) reconcileWorkspaces(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	enabledWorkspaces := instance.Spec.Workspaces.Enabled

	if enabledWorkspaces {
		if result, err := r.addWorkspaces(instance, users, appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addWorkspaces(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	reqLogger := log.WithName("Workspaces")
	enabledOpenShiftoAuth := instance.Spec.Workspaces.OpenShiftoAuth

	workspacesNamespace := deployment.NewNamespace(instance, "workspaces")
	if err := r.client.Create(context.TODO(), workspacesNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Workspaces Projects")
	}

	workspacesCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "checlusters.org.eclipse.che", "org.eclipse.che", "CheCluster", "CheClusterList", "checlusters", "checluster", "v1", nil, nil)
	if err := r.client.Create(context.TODO(), workspacesCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Workspaces Custom Resource Definition")
	}

	workspacesServiceAccount := deployment.NewServiceAccount(instance, "workspaces-operator", workspacesNamespace.Name)
	if err := r.client.Create(context.TODO(), workspacesServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Workspaces Service Account")
	}

	workspacesClusterRole := deployment.NewClusterRole(instance, "workspaces-operator", workspacesNamespace.Name, deployment.WorkspacesRules())
	if err := r.client.Create(context.TODO(), workspacesClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Workspaces Cluster Role")
	}

	workspacesClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "workspaces-operator", workspacesNamespace.Name, "workspaces-operator", "workspaces-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), workspacesClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Workspaces Cluster Role Binding")
	}

	commands := []string{
		"che-operator",
	}
	workspacesOperator := deployment.NewOperatorDeployment(instance, "workspaces-operator", workspacesNamespace.Name, "registry.redhat.io/codeready-workspaces/server-operator-rhel8:1.2", "workspaces-operator", 60000, commands, nil, nil, nil)
	if err := r.client.Create(context.TODO(), workspacesOperator); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Workspaces Operator")
	}

	workspacesCustomResource := checlustercustomresource.NewCheClusterCustomResource(instance, "codeready", workspacesNamespace.Name, "registry.redhat.io/codeready-workspaces/server-rhel8", "1.2", false, false, enabledOpenShiftoAuth)
	if err := r.client.Create(context.TODO(), workspacesCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Workspaces Custom Resource")
	}

	openshiftStackImageStream := deployment.NewImageStream(instance, "che-cloud-native", "openshift", "quay.io/mcouliba/che-cloud-native:ocp4", "ocp4")
	if err := r.client.Create(context.TODO(), openshiftStackImageStream); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Cloud Native Stack Image Stream (OCP4)")
	}

	var (
		body           []byte
		err            error
		url            string
		httpResponse   *http.Response
		httpRequest    *http.Request
		codereadyToken util.Token
		masterToken    util.Token
		client         = &http.Client{}
		stackResponse  = codereadystack.Stack{}
		timeout        = 100
	)

	// Wait for CodeReady Workspaces to be running
	logrus.Infof("Waiting for CodeReady Workspaces Operator to build resources (%v seconds)", timeout)
	time.Sleep(time.Duration(timeout) * time.Second)

	workspacesDeployment, err := r.GetEffectiveDeployment(instance, "codeready", workspacesNamespace.Name)
	if err != nil {
		reqLogger.Error(err, "Failed to get codeready deployment")
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	}

	if workspacesDeployment.Status.AvailableReplicas != 1 {
		scaled := k8sclient.GetDeploymentStatus("codeready", workspacesNamespace.Name)
		if !scaled {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, err
		}
	}

	url = "http://keycloak-workspaces." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"
	httpRequest, err = http.NewRequest("POST", url, strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
			return reconcile.Result{}, err
		}
	}

	if !enabledOpenShiftoAuth {
		openshiftUserPassword := instance.Spec.UserPassword
		for id := 1; id <= users; id++ {
			username := fmt.Sprintf("user%d", id)
			body, err = json.Marshal(codereadyuser.NewCodeReadyUser(instance, username, openshiftUserPassword))
			if err != nil {
				return reconcile.Result{}, err
			}

			httpRequest, err = http.NewRequest("POST", "http://keycloak-workspaces."+appsHostnameSuffix+"/auth/admin/realms/codeready/users", bytes.NewBuffer(body))
			httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)
			httpRequest.Header.Set("Content-Type", "application/json")

			httpResponse, err = client.Do(httpRequest)
			if err != nil {
				reqLogger.Info("Error when creating " + username + " for CodeReady Workspaces")
				return reconcile.Result{}, err
			}
			if httpResponse.StatusCode == http.StatusCreated {
				reqLogger.Info("Created " + username + " for CodeReady Workspaces")
			}
		}
	}

	httpRequest, err = http.NewRequest("POST", "http://keycloak-workspaces."+appsHostnameSuffix+"/auth/realms/codeready/protocol/openid-connect/token", strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		reqLogger.Info("Error when getting Workspaces Access Token")
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&codereadyToken); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Workspaces Factory
	body, err = json.Marshal(codereadyfactory.NewDebuggingFactory(openshiftConsoleURL, openshiftAPIURL, appsHostnameSuffix, instance.Spec.UserPassword))
	if err != nil {
		return reconcile.Result{}, err
	}

	httpRequest, err = http.NewRequest("POST", "http://codeready-workspaces."+appsHostnameSuffix+"/api/factory", bytes.NewBuffer(body))
	httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusCreated || httpResponse.StatusCode == http.StatusOK {
		reqLogger.Info("Created Debugging Factory")
	}

	body, err = json.Marshal(codereadystack.NewCloudNativeStack(instance))
	if err != nil {
		return reconcile.Result{}, err
	}

	httpRequest, err = http.NewRequest("POST", "http://codeready-workspaces."+appsHostnameSuffix+"/api/stack", bytes.NewBuffer(body))
	httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusCreated || httpResponse.StatusCode == http.StatusOK {
		reqLogger.Info("Created Cloud Native Stack")

		if err := json.NewDecoder(httpResponse.Body).Decode(&stackResponse); err != nil {
			return reconcile.Result{}, err
		}

		body, err = json.Marshal(codereadystack.NewCloudNativeStackPermission(instance, stackResponse.ID))
		if err != nil {
			return reconcile.Result{}, err
		}

		httpRequest, err = http.NewRequest("POST", "http://codeready-workspaces."+appsHostnameSuffix+"/api/permissions", bytes.NewBuffer(body))
		httpRequest.Header.Set("Authorization", "Bearer "+codereadyToken.AccessToken)
		httpRequest.Header.Set("Content-Type", "application/json")

		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			return reconcile.Result{}, err
		}
		if httpResponse.StatusCode == http.StatusCreated {
			reqLogger.Info("Granted Cloud Native Stack")
		}

	}

	//Success
	return reconcile.Result{}, nil
}
