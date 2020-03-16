package workshop

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

// Reconciling CodeReadyWorkspace
func (r *ReconcileWorkshop) reconcileCodeReadyWorkspace(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	enabledCodeReadyWorkspace := instance.Spec.Infrastructure.CodeReadyWorkspace.Enabled

	if enabledCodeReadyWorkspace {

		if result, err := r.addCodeReadyWorkspace(instance, users, appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.CodeReadyWorkspace != util.OperatorStatus.Installed {
			instance.Status.CodeReadyWorkspace = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addCodeReadyWorkspace(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {

	channel := instance.Spec.Infrastructure.CodeReadyWorkspace.OperatorHub.Channel
	clusterServiceVersion := instance.Spec.Infrastructure.CodeReadyWorkspace.OperatorHub.ClusterServiceVersion

	codeReadyWorkspacesNamespace := deployment.NewNamespace(instance, "workspaces")
	if err := r.client.Create(context.TODO(), codeReadyWorkspacesNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", codeReadyWorkspacesNamespace.Name)
	}

	codeReadyWorkspacesOperatorGroup := deployment.NewOperatorGroup(instance, "codeready-workspaces", codeReadyWorkspacesNamespace.Name)
	if err := r.client.Create(context.TODO(), codeReadyWorkspacesOperatorGroup); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OperatorGroup", codeReadyWorkspacesOperatorGroup.Name)
	}

	codeReadyWorkspacesSubscription := deployment.NewRedHatSubscription(instance, "codeready-workspaces", codeReadyWorkspacesNamespace.Name,
		"codeready-workspaces", channel, clusterServiceVersion)
	if err := r.client.Create(context.TODO(), codeReadyWorkspacesSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", codeReadyWorkspacesSubscription.Name)
	}

	// Approve the installation
	if err := r.ApproveInstallPlan(clusterServiceVersion, "codeready-workspaces", codeReadyWorkspacesNamespace.Name); err != nil {
		logrus.Warnf("Waiting for Subscription to create InstallPlan for %s", "codeready-workspaces")
		return reconcile.Result{}, err
	}

	// Wait for CodeReadyWorkspace Operator to be running
	if !k8sclient.GetDeploymentStatus("codeready-operator", codeReadyWorkspacesNamespace.Name) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, nil
	}

	codeReadyWorkspacesCustomResource := deployment.NewCodeReadyWorkspacesCustomResource(instance, "codereadyworkspaces", codeReadyWorkspacesNamespace.Name)
	if err := r.client.Create(context.TODO(), codeReadyWorkspacesCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", codeReadyWorkspacesCustomResource.Name)
	}

	// Wait for CodeReadyWorkspace to be running
	if !k8sclient.GetDeploymentStatus("codeready", codeReadyWorkspacesNamespace.Name) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, nil
	}

	// Initialize Workspaces from devfile
	devfile, result, err := getDevFile(instance)
	if err != nil {
		return result, err
	}

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)

		userAccessToken, result, err := getUserToken(instance, username, "codeready", codeReadyWorkspacesNamespace.Name, appsHostnameSuffix)
		if err != nil {
			return result, err
		}

		if result, err := updateUserEmail(instance, username, "codeready", codeReadyWorkspacesNamespace.Name, appsHostnameSuffix); err != nil {
			return result, err
		}

		if result, err := initWorkspace(instance, username, "codeready", codeReadyWorkspacesNamespace.Name, userAccessToken, devfile, appsHostnameSuffix); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func getDevFile(instance *openshiftv1alpha1.Workshop) (string, reconcile.Result, error) {

	var (
		httpResponse *http.Response
		httpRequest  *http.Request
		devfile      string
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

	gitURL, err := url.Parse(instance.Spec.Source.GitURL)
	if err != nil {
		return "", reconcile.Result{}, err
	}
	devfileRawURL := fmt.Sprintf("https://raw.githubusercontent.com%s/%s/devfile.yaml", gitURL.Path, instance.Spec.Source.GitBranch)
	httpRequest, err = http.NewRequest("GET", devfileRawURL, nil)

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		logrus.Errorf("Error when getting Devfile from %s", devfileRawURL)
		return "", reconcile.Result{}, err
	}

	if httpResponse.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
		if err != nil {
			logrus.Errorf("Error when reading %s", devfileRawURL)
			return "", reconcile.Result{}, err
		}

		bodyJSON, err := yaml.YAMLToJSON(bodyBytes)
		if err != nil {
			logrus.Errorf("Error to converting %s to JSON", devfileRawURL)
			return "", reconcile.Result{}, err
		}
		devfile = string(bodyJSON)
	} else {
		logrus.Errorf("Error (%v) when getting Devfile from %s", httpResponse.StatusCode, devfileRawURL)
		return "", reconcile.Result{}, err
	}

	return devfile, reconcile.Result{}, nil
}

func getUserToken(instance *openshiftv1alpha1.Workshop, username string,
	codeflavor string, namespace string, appsHostnameSuffix string) (string, reconcile.Result, error) {
	var (
		openshiftUserPassword = instance.Spec.User.Password
		err                   error
		httpResponse          *http.Response
		httpRequest           *http.Request
		keycloakCheTokenURL   = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/realms/" + codeflavor + "/protocol/openid-connect/token"
		oauthOpenShiftURL     = "https://oauth-openshift." + appsHostnameSuffix + "/oauth/authorize?client_id=openshift-challenging-client&response_type=token"

		userToken util.Token
		client    = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	// GET TOKEN
	httpRequest, err = http.NewRequest("GET", oauthOpenShiftURL, nil)
	httpRequest.Header.Set("Authorization", "Basic "+util.GetBasicAuth(username, openshiftUserPassword))
	httpRequest.Header.Set("X-CSRF-Token", "xxx")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		logrus.Errorf("Error when getting Token Exchange for %s: %v", username, err)
		return "", reconcile.Result{}, err
	}

	if httpResponse.StatusCode == http.StatusFound {
		locationURL, err := url.Parse(httpResponse.Header.Get("Location"))
		if err != nil {
			return "", reconcile.Result{}, err
		}

		regex := regexp.MustCompile("access_token=([^&]+)")
		subjectToken := regex.FindStringSubmatch(locationURL.Fragment)

		// Get User Access Token
		data := url.Values{}
		data.Set("client_id", codeflavor+"-public")
		data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
		data.Set("subject_token", subjectToken[1])
		data.Set("subject_issuer", "openshift-v4")
		data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")

		httpRequest, err = http.NewRequest("POST", keycloakCheTokenURL, strings.NewReader(data.Encode()))
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			logrus.Errorf("Error to get the user access  token from %s keycloak (%v)", codeflavor, err)
			return "", reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusOK {
			if err := json.NewDecoder(httpResponse.Body).Decode(&userToken); err != nil {
				logrus.Errorf("Error to get the user access  token from %s keycloak (%v)", codeflavor, err)
				return "", reconcile.Result{}, err
			}
		} else {
			logrus.Errorf("Error to get the user access token from %s keycloak (%d)", codeflavor, httpResponse.StatusCode)
			return "", reconcile.Result{}, err
		}
	} else {
		logrus.Errorf("Error when getting Token Exchange for %s (%d)", username, httpResponse.StatusCode)
		return "", reconcile.Result{}, err
	}

	return userToken.AccessToken, reconcile.Result{}, nil
}

func updateUserEmail(instance *openshiftv1alpha1.Workshop, username string,
	codeflavor string, namespace string, appsHostnameSuffix string) (reconcile.Result, error) {
	var (
		err                    error
		httpResponse           *http.Response
		httpRequest            *http.Request
		keycloakMasterTokenURL = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"
		keycloakUserURL        = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/admin/realms/" + codeflavor + "/users"
		masterToken            util.Token
		client                 = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		cheUser []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
	)

	// Get Keycloak Admin Token
	httpRequest, err = http.NewRequest("POST", keycloakMasterTokenURL, strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		logrus.Errorf("Error when getting the master token from %s keycloak (%v)", codeflavor, err)
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
			logrus.Errorf("Error when reading the master token: %v", err)
			return reconcile.Result{}, err
		}
	} else {
		logrus.Errorf("Error when getting the master token from %s keycloak (%d)", codeflavor, httpResponse.StatusCode)
		return reconcile.Result{}, err
	}

	// GET USER
	httpRequest, err = http.NewRequest("GET", keycloakUserURL+"?username="+username, nil)
	httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		logrus.Errorf("Error when getting %s user: %v", username, err)
		return reconcile.Result{}, err
	}

	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&cheUser); err != nil {
			logrus.Errorf("Error to get the user info (%v)", err)
			return reconcile.Result{}, err
		}

		if cheUser[0].Email == "" {
			httpRequest, err = http.NewRequest("PUT", keycloakUserURL+"/"+cheUser[0].ID,
				strings.NewReader(`{"email":"`+username+`@none.com"}`))
			httpRequest.Header.Set("Content-Type", "application/json")
			httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)
			httpResponse, err = client.Do(httpRequest)
			if err != nil {
				logrus.Errorf("Error when update email address for %s: %v", username, err)
				return reconcile.Result{}, err
			}
		}
	} else {
		logrus.Errorf("Error when getting %s user: %v", username, httpResponse.StatusCode)
		return reconcile.Result{}, err
	}

	//Success
	return reconcile.Result{}, nil
}

func initWorkspace(instance *openshiftv1alpha1.Workshop, username string,
	codeflavor string, namespace string, userAccessToken string, devfile string,
	appsHostnameSuffix string) (reconcile.Result, error) {

	var (
		err                 error
		httpResponse        *http.Response
		httpRequest         *http.Request
		devfileWorkspaceURL = "https://" + codeflavor + "-" + namespace + "." + appsHostnameSuffix + "/api/workspace/devfile?start-after-create=true&namespace=" + username

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

	httpRequest, err = http.NewRequest("POST", devfileWorkspaceURL, strings.NewReader(devfile))
	httpRequest.Header.Set("Authorization", "Bearer "+userAccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		logrus.Errorf("Error when creating the workspace for %s: %v", username, err)
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	// if httpResponse.StatusCode == http.StatusOK {
	// 	logrus.Infof("Started Workspace for %s", username)
	// } else {
	// 	logrus.Errorf("Error (%d) when creating the workspace for %s", httpResponse.StatusCode, username)
	// 	return reconcile.Result{}, err
	// }

	//Success
	return reconcile.Result{}, nil
}
