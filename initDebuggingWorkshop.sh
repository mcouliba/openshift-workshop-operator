#!/bin/bash

GUID=$1
OPENSHIFT_API_URL="https://api.cluster-${GUID}.${GUID}.events.opentlc.com:6443"
KEYCLOAK_MASTER_URL="http://keycloak-eclipse-che.apps.cluster-${GUID}.${GUID}.events.opentlc.com/auth/realms/master/protocol/openid-connect/token"
KEYCLOAK_CHE_URL="http://keycloak-eclipse-che.apps.cluster-${GUID}.${GUID}.events.opentlc.com/auth/realms/che/protocol/openid-connect/token"
KEYCLOAK_USER_URL="http://keycloak-eclipse-che.apps.cluster-${GUID}.${GUID}.events.opentlc.com/auth/admin/realms/che/users"
WORKSHOP_DEVFILE_URL="https://raw.githubusercontent.com/mcouliba/debugging-workshop/master/devfile.yaml"
WORKSHOP_DEVFILE=$(curl -s  ${WORKSHOP_DEVFILE_URL} | yaml2json)
CHE_URL="http://che-eclipse-che.apps.cluster-${GUID}.${GUID}.events.opentlc.com"
WORKSPACE_URL="${CHE_URL}/api/workspace"
DEBUGGING_WORKSPACE_URL="${WORKSPACE_URL}/%3Awksp-debugging?includeInternalServers=false"
DEVFILE_WORKSPACE_URL="${WORKSPACE_URL}/devfile?start-after-create=true"

# for i in {1..18};
for i in {18..35};
do
    
    che_user="user$i"
    subject_token=$(oc login -u "$che_user" -p 'r3dh4t1!' "${OPENSHIFT_API_URL}" &> /dev/null &&\
        oc whoami --show-token)

    # echo ">>>>> Subject Token: $subject_token"
        
    access_token=$(curl -s  -X POST  \
        -d "client_id=che-public" \
        --data-urlencode "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
        -d "subject_token=${subject_token}" \
        -d "subject_issuer=openshift-v4" \
        --data-urlencode "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
      ${KEYCLOAK_CHE_URL} | jq -r '.access_token')

    # echo ">>>>> Access Token: $access_token"

    if [ ! -z "$access_token" ];
    then 
        
        # Start Workspaces
        master_access_token=$(curl -s  -X POST  \
                -d "client_id=admin-cli" \
                -d "username=admin" \
                -d "password=admin" \
                -d "grant_type=password" \
                ${KEYCLOAK_MASTER_URL} | jq -r '.access_token')

        # echo ">>>>> Master Access Token: $master_access_token"

        userid=$(curl -s  -X GET \
            --header "Authorization: Bearer $master_access_token"\
            "${KEYCLOAK_USER_URL}?username=${che_user}" | jq -r '.[0].id')
        
        # echo ">>>>> User ID: ${userid}"
    
        curl -s -X PUT "${KEYCLOAK_USER_URL}/${userid}" \
            --header "Content-Type: application/json" \
            --header "Authorization: Bearer $master_access_token" \
            -d "{\"email\":\"${che_user}@none.com\"}"

        
        
        curl -s -X POST  "${DEVFILE_WORKSPACE_URL}&namespace=$che_user" \
            --header "Content-Type: application/json" \
            --header "Accept: application/json" \
            --header "Authorization: Bearer $access_token" \
            -d "${WORKSHOP_DEVFILE}"

        workspaceid=$(curl -s  -X GET \
            --header "Authorization: Bearer $access_token"\
            "${DEBUGGING_WORKSPACE_URL}" | jq -r '.id')

        echo ">>>>> Creating Eclipse Che ${workspaceid} for user $che_user"

        if [ ! -z "$workspaceid" ];
        then 
            # Grant Squash Role to SA
            oc login -u "opentlc-mgr" -p 'r3dh4t1!' "${OPENSHIFT_API_URL}" &> /dev/null
            oc adm policy add-cluster-role-to-user cluster-admin -z che-workspace -n $workspaceid
        fi
    fi
done

