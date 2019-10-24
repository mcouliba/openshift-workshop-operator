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

	"github.com/ghodss/yaml"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	che "github.com/redhat/openshift-workshop-operator/pkg/deployment/che"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Che
func (r *ReconcileWorkshop) reconcileChe(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	enabledChe := instance.Spec.Infrastructure.Che.Enabled

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

	cheClusterServiceVersion := instance.Spec.Infrastructure.Che.OperatorHub.ClusterServiceVersion

	cheNamespace := deployment.NewNamespace(instance, "eclipse-che")
	if err := r.client.Create(context.TODO(), cheNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", cheNamespace.Name)
	}

	// cheCatalogSourceConfig := deployment.NewCatalogSourceConfig(instance, "installed-eclipse-che", cheNamespace.Name, "eclipse-che")
	// if err := r.client.Create(context.TODO(), cheCatalogSourceConfig); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	logrus.Infof("Created %s CatalogSourceConfig", cheCatalogSourceConfig.Name)
	// }

	cheOperatorGroup := deployment.NewOperatorGroup(instance, "eclipse-che", cheNamespace.Name)
	if err := r.client.Create(context.TODO(), cheOperatorGroup); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OperatorGroup", cheOperatorGroup.Name)
	}

	cheSubscription := deployment.NewCommunitySubscription(instance, "eclipse-che", cheNamespace.Name, "eclipse-che",
		instance.Spec.Infrastructure.Che.OperatorHub.Channel,
		cheClusterServiceVersion)
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
		timeout = 120
	)

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

	// Wait for Che to be running
	cheDeployment, err := r.GetEffectiveDeployment(instance, "che", cheNamespace.Name)
	if err != nil {
		logrus.Infof("Waiting for Che Operator to build resources (%v seconds)", timeout)
		time.Sleep(time.Duration(timeout) * time.Second)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, err
	}

	if cheDeployment.Status.AvailableReplicas != 1 {
		scaled := k8sclient.GetDeploymentStatus("che", cheNamespace.Name)
		if !scaled {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, err
		}
	}

	// Initialize Workspaces from devfile
	devfile, result, err := getDevFile(instance)
	if err != nil {
		return result, err
	}

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)

		userAccessToken, result, err := getUserToken(instance, username, cheNamespace.Name, appsHostnameSuffix)
		if err != nil {
			return result, err
		}

		if result, err := updateUserEmail(instance, username, cheNamespace.Name, appsHostnameSuffix); err != nil {
			return result, err
		}

		if result, err := initWorkspace(instance, username, userAccessToken, devfile, appsHostnameSuffix); err != nil {
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
	cheNamespace string, appsHostnameSuffix string) (string, reconcile.Result, error) {
	var (
		openshiftUserPassword = instance.Spec.User.Password
		err                   error
		httpResponse          *http.Response
		httpRequest           *http.Request
		keycloakCheTokenURL   = "http://keycloak-" + cheNamespace + "." + appsHostnameSuffix + "/auth/realms/che/protocol/openid-connect/token"
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
		data.Set("client_id", "che-public")
		data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
		data.Set("subject_token", subjectToken[1])
		data.Set("subject_issuer", "openshift-v4")
		data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")

		httpRequest, err = http.NewRequest("POST", keycloakCheTokenURL, strings.NewReader(data.Encode()))
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			logrus.Errorf("Error to get the user access  token from che keycloak (%v)", err)
			return "", reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusOK {
			if err := json.NewDecoder(httpResponse.Body).Decode(&userToken); err != nil {
				logrus.Errorf("Error to get the user access  token from che keycloak (%v)", err)
				return "", reconcile.Result{}, err
			}
		} else {
			logrus.Errorf("Error to get the user access token from che keycloak (%d)", httpResponse.StatusCode)
			return "", reconcile.Result{}, err
		}
	} else {
		logrus.Errorf("Error when getting Token Exchange for %s (%d)", username, httpResponse.StatusCode)
		return "", reconcile.Result{}, err
	}

	return userToken.AccessToken, reconcile.Result{}, nil
}

func updateUserEmail(instance *openshiftv1alpha1.Workshop, username string,
	cheNamespace string, appsHostnameSuffix string) (reconcile.Result, error) {
	var (
		err                    error
		httpResponse           *http.Response
		httpRequest            *http.Request
		keycloakMasterTokenURL = "http://keycloak-" + cheNamespace + "." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"
		keycloakUserURL        = "http://keycloak-" + cheNamespace + "." + appsHostnameSuffix + "/auth/admin/realms/che/users"
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
		logrus.Errorf("Error when getting the master token from che keycloak (%v)", err)
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
			logrus.Errorf("Error when reading the master token: %v", err)
			return reconcile.Result{}, err
		}
	} else {
		logrus.Errorf("Error when getting the master token from che keycloak (%d)", httpResponse.StatusCode)
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
	userAccessToken string, devfile string,
	appsHostnameSuffix string) (reconcile.Result, error) {

	var (
		err                 error
		httpResponse        *http.Response
		httpRequest         *http.Request
		devfileWorkspaceURL = "http://che-eclipse-che." + appsHostnameSuffix + "/api/workspace/devfile?start-after-create=true&namespace=" + username

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
