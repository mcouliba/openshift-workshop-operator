package workshop

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

	namespace := deployment.NewNamespace(instance, "argocd")
	if err := r.client.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", namespace.Name)
	}

	catalogSource := deployment.NewCatalogSource(instance, "argocd-catalog", "quay.io/mcouliba/argocd-operator-registry:latest", "Argo CD Operator", "Argo CD")
	if err := r.client.Create(context.TODO(), catalogSource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Catalog Source", catalogSource.Name)
	}

	operatorGroup := deployment.NewOperatorGroup(instance, "argocd-operator", namespace.Name)
	if err := r.client.Create(context.TODO(), operatorGroup); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OperatorGroup", operatorGroup.Name)
	}

	subscription := deployment.NewCustomSubscription(instance, "argocd-operator", namespace.Name, "argocd-operator",
		"alpha", catalogSource.Name)
	if err := r.client.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", subscription.Name)
	}

	// Approve the installation
	if err := r.ApproveInstallPlan("argocd-operator", namespace.Name); err != nil {
		logrus.Warnf("Waiting for Subscription to create InstallPlan for %s", "argocd-operator")
		return reconcile.Result{}, err
	}

	// Wait for ArgoCD Operator to be running
	time.Sleep(time.Duration(1) * time.Second)
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

	argocdRoute := "https://argocd-server-argocd." + appsHostnameSuffix
	redirectURIs := []string{argocdRoute + "/api/dex/callback"}
	oauthClient := deployment.NewOAuthClient(instance, "argocd-dex", redirectURIs)
	if err := r.client.Create(context.TODO(), oauthClient); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OAuth Client", oauthClient.Name)
	}

	data := map[string]string{
		"dex.config": `connectors:
  - type: openshift
    id: openshift
    name: OpenShift
    config:
      issuer: ` + openshiftAPIURL + `:6443
      clientID: ` + oauthClient.Name + `
      clientSecret: openshift
      insecureCA: true
      redirectURI: ` + argocdRoute + `/api/dex/callback`,
		"url":                          argocdRoute,
		"application.instanceLabelKey": "argocd.argoproj.io/instance",
	}

	labels := map[string]string{
		"app.kubernetes.io/name":    "argocd-cm",
		"app.kubernetes.io/part-of": "argocd",
	}

	configmap := deployment.NewConfigMap(instance, "argocd-cm", namespace.Name, labels, data)
	if err := r.client.Create(context.TODO(), configmap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s ConfigMap", configmap.Name)
	} else if errors.IsAlreadyExists(err) {
		configMapFound := &corev1.ConfigMap{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: configmap.Name, Namespace: namespace.Name}, configMapFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(configMapFound.Data, data) {
				// Update
				configMapFound.Data = data
				if err := r.client.Update(context.TODO(), configMapFound); err != nil {
					return reconcile.Result{}, err
				}
				logrus.Infof("Updated %s ConfigMap", configMapFound.Name)
			}
		}
	}

	argocdPolicy := ""
	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		userRole := fmt.Sprintf("role:%s", username)
		projectName := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.StagingName, id)

		userPolicy := `p, ` + userRole + `, applications, create, ` + projectName + `/*, allow
p, ` + userRole + `, applications, get, default/` + projectName + `, allow
p, ` + userRole + `, applications, sync, default/` + projectName + `, allow
p, ` + userRole + `, clusters, get, https://kubernetes.default.svc, allow
p, ` + userRole + `, projects, get, default, allow
p, ` + userRole + `, repositories, get, ` + instance.Spec.Source.GitURL + `, allow
g, ` + username + `, ` + userRole + `
`
		argocdPolicy = fmt.Sprintf("%s%s", argocdPolicy, userPolicy)
	}

	rbacData := map[string]string{
		"scopes":     "[preferred_username]",
		"policy.csv": argocdPolicy,
	}

	rbacLabels := map[string]string{
		"app.kubernetes.io/name":    "argocd-rbac-cm",
		"app.kubernetes.io/part-of": "argocd",
	}

	rbacConfigmap := deployment.NewConfigMap(instance, "argocd-rbac-cm", namespace.Name, rbacLabels, rbacData)
	if err := r.client.Create(context.TODO(), rbacConfigmap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s ConfigMap", rbacConfigmap.Name)
	} else if errors.IsAlreadyExists(err) {
		configMapFound := &corev1.ConfigMap{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: rbacConfigmap.Name, Namespace: namespace.Name}, configMapFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(configMapFound.Data, rbacData) {
				// Update
				configMapFound.Data = rbacData
				if err := r.client.Update(context.TODO(), configMapFound); err != nil {
					return reconcile.Result{}, err
				}
				logrus.Infof("Updated %s ConfigMap", configMapFound.Name)
			}
		}
	}

	customResource := deployment.NewArgoCDCustomResource(instance, "argocd", namespace.Name)
	if err := r.client.Create(context.TODO(), customResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", customResource.Name)
	}

	// Wait for ArgoCD Server to be running
	time.Sleep(time.Duration(1) * time.Second)
	deployment, err := r.GetEffectiveDeployment(instance, "argocd-server", namespace.Name)
	if err != nil {
		logrus.Warnf("Waiting for %s to be running", "argocd-server")
		return reconcile.Result{}, err
	}

	if deployment.Status.AvailableReplicas != 1 {
		scaled := k8sclient.GetDeploymentStatus("argocd-server", namespace.Name)
		if !scaled {
			return reconcile.Result{}, err
		}
	}

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
