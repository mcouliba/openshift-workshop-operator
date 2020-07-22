package workshop

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling ArgoCD
func (r *ReconcileWorkshop) reconcileArgoCD(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	enabledArgoCD := instance.Spec.Infrastructure.ArgoCD.Enabled

	if enabledArgoCD {

		if result, err := r.addArgoCD(instance, users, appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.ArgoCD != util.OperatorStatus.Installed {
			instance.Status.ArgoCD = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addArgoCD(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {

	channel := instance.Spec.Infrastructure.ArgoCD.OperatorHub.Channel
	clusterServiceVersion := instance.Spec.Infrastructure.ArgoCD.OperatorHub.ClusterServiceVersion

	namespace := deployment.NewNamespace(instance, "argocd")
	if err := r.client.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", namespace.Name)
	}

	operatorGroup := deployment.NewOperatorGroup(instance, "argocd-operator", namespace.Name)
	if err := r.client.Create(context.TODO(), operatorGroup); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OperatorGroup", operatorGroup.Name)
	}

	subscription := deployment.NewCommunitySubscription(instance, "argocd-operator", namespace.Name,
		"argocd-operator", channel, clusterServiceVersion)
	if err := r.client.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", subscription.Name)
	}

	// Approve the installation
	if err := r.ApproveInstallPlan(clusterServiceVersion, "argocd-operator", namespace.Name); err != nil {
		logrus.Warnf("Waiting for Subscription to create InstallPlan for %s", "argocd-operator")
		return reconcile.Result{}, err
	}

	// Wait for ArgoCD Operator to be running
	operatorDeployment, err := r.GetEffectiveDeployment(instance, "argocd-operator", namespace.Name)
	if err != nil {
		logrus.Warnf("Waiting for OLM to create %s deployment", "argocd-operator")
		return reconcile.Result{}, err
	}

	if operatorDeployment.Status.AvailableReplicas != 1 {
		scaled := k8sclient.GetDeploymentStatus("argocd-operator", namespace.Name)
		if !scaled {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, err
		}
	}

	argocdPolicy := ""
	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		userRole := fmt.Sprintf("role:%s", username)
		projectName := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.StagingName, id)

		userPolicy := `p, ` + userRole + `, applications, *, default/` + projectName + `, allow
p, ` + userRole + `, clusters, get, https://kubernetes.default.svc, allow
p, ` + userRole + `, projects, get, default, allow
p, ` + userRole + `, repositories, get, ` + instance.Spec.Source.GitURL + `, allow
p, ` + userRole + `, repositories, get, http://gogs-gogs-server.workshop-infra.svc:3000/` + username + `/gitops-cn-project.git, allow
p, ` + userRole + `, repositories, create, http://gogs-gogs-server.workshop-infra.svc:3000/` + username + `/gitops-cn-project.git, allow
p, ` + userRole + `, repositories, delete, http://gogs-gogs-server.workshop-infra.svc:3000/` + username + `/gitops-cn-project.git, allow
g, ` + username + `, ` + userRole + `
`
		argocdPolicy = fmt.Sprintf("%s%s", argocdPolicy, userPolicy)
	}

	customResource := deployment.NewArgoCDCustomResource(instance, "argocd", namespace.Name, argocdPolicy)
	if err := r.client.Create(context.TODO(), customResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", customResource.Name)
	}

	// Wait for ArgoCD Dex Server to be running
	if !k8sclient.GetDeploymentStatus("argocd-dex-server", namespace.Name) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, nil
	}

	// Wait for ArgoCD Server to be running
	if !k8sclient.GetDeploymentStatus("argocd-server", namespace.Name) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, nil
	}

	time.Sleep(time.Duration(10) * time.Second)
	adminToken, result, err := getAdminToken(instance, namespace.Name, appsHostnameSuffix)
	if err != nil {
		return result, err
	}

	if result, err := createRepository(instance, adminToken, appsHostnameSuffix); err != nil {
		return result, err
	}

	for id := 1; id <= users; id++ {
		stagingProject := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.StagingName, id)

		if result, err := createApplication(instance, stagingProject, adminToken, appsHostnameSuffix); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func getAdminToken(instance *openshiftv1alpha1.Workshop, namespace string, appsHostnameSuffix string) (string, reconcile.Result, error) {
	var (
		err          error
		httpResponse *http.Response
		httpRequest  *http.Request
		sessionURL   = "https://argocd-server-argocd." + appsHostnameSuffix + "/api/v1/session"

		adminToken util.ArgoToken
		client     = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	serverPodname, err := k8sclient.GetDeploymentPod("argocd-server", namespace, "app.kubernetes.io/name")
	if err == nil {
		body := "{\"username\": \"admin\", \"password\":\"" + serverPodname + "\"}"

		// GET TOKEN
		httpRequest, err = http.NewRequest("POST", sessionURL, strings.NewReader(body))

		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			logrus.Errorf("Error when getting Argo CD token: %v", err)
			return "", reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK {
			if err := json.NewDecoder(httpResponse.Body).Decode(&adminToken); err != nil {
				logrus.Errorf("Error when parsing Argo CD Token: %v", err)
				return "", reconcile.Result{}, err
			}
		} else {
			logrus.Errorf("Error when getting Argo CD token (%d)", httpResponse.StatusCode)
			return "", reconcile.Result{}, err
		}
	} else {
		logrus.Errorf("Error when getting Argo CD Server Pod Name: %v", err)
		return "", reconcile.Result{}, err
	}
	return adminToken.Token, reconcile.Result{}, nil
}

func createRepository(instance *openshiftv1alpha1.Workshop, token string,
	appsHostnameSuffix string) (reconcile.Result, error) {

	var (
		err          error
		httpResponse *http.Response
		httpRequest  *http.Request
		repoURL      = "https://argocd-server-argocd." + appsHostnameSuffix + "/api/v1/repositories"
		body         = "{\"repo\": \"" + instance.Spec.Source.GitURL + "\"}"
		client       = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	httpRequest, err = http.NewRequest("POST", repoURL, strings.NewReader(body))
	httpRequest.Header.Set("Authorization", "Bearer "+token)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		logrus.Errorf("Error when creating a Argo CD Repository %s: %v", instance.Spec.Source.GitURL, err)
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()

	//Success
	return reconcile.Result{}, nil
}

func createApplication(instance *openshiftv1alpha1.Workshop, stagingProject string, token string,
	appsHostnameSuffix string) (reconcile.Result, error) {

	var (
		err          error
		httpResponse *http.Response
		httpRequest  *http.Request
		repoURL      = "https://argocd-server-argocd." + appsHostnameSuffix + "/api/v1/applications"
		body         = `{
"metadata": {
		"name": "` + stagingProject + `"
	},
	"spec": {
		"source": {
			"repoURL": "` + instance.Spec.Source.GitURL + `",
			"path": "gitops",
			"targetRevision": "` + instance.Spec.Source.GitBranch + `"
		},
		"destination": {
			"server": "https://kubernetes.default.svc",
			"namespace": "` + stagingProject + `"
		},
		"project": "default"
	}
}`
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	httpRequest, err = http.NewRequest("POST", repoURL, strings.NewReader(body))
	httpRequest.Header.Set("Authorization", "Bearer "+token)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		logrus.Errorf("Error when creating a Argo CD Application for %s: %v", stagingProject, err)
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()

	//Success
	return reconcile.Result{}, nil
}
