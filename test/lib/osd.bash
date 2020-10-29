
TMPDIR="${BATS_TMPDIR:-/tmp}"
KIND_CLUSTER="${KIND_CLUSTER:-lpabon-kind-csi}"


# Only show output of the program on failure
function osd::suppress() {
    (
        local output=/tmp/output.$$
        rm --force ${output} 2> /dev/null
        ${1+"$@"} > ${output} 2>&1
        result=$?
        if [ $result -ne 0 ] ; then
            cat ${output}
        fi
        rm ${output}
        exit $result
    )
}

# TAP message
function osd::echo() {
    if [ $DEBUG -eq 1 ] ; then
        echo "# ${1}" >&3
    fi
}

# TAP compliant steps which can be printed out
function osd::by() {
    if [ $DEBUG -eq 1 ] ; then
        echo "# STEP: ${1}" >&3
    fi
}

# Get the Kind cluster IP from docker
function osd::clusterip() {
    docker inspect ${CLUSTER_CONTROL_PLANE_CONTAINER} | jq -r '.[].NetworkSettings.Networks.kind.IPAddress'
}

# Return the SDK REST Gateway address
function osd::getSdkRestGWEndpoint() {
    local clusterip=$(osd::clusterip)
    local nodeport=$(kubectl -n kube-system get svc portworx-api -o json | jq '.spec.ports[2].nodePort')
    echo ${clusterip}:${nodeport}
}

# Return the SDK gRPC endpoint
function osd::getSdkEndpoint() {
    local clusterip=$(osd::clusterip)
    local nodeport=$(kubectl -n kube-system get svc portworx-api -o json | jq '.spec.ports[1].nodePort')
    echo ${clusterip}:${nodeport}
}

# Creats a user in Kubernetes only. Use osd::createUserKubeconfig() instead to create a full
# kubeconfig for the new user.
function osd::createUser() {
    local username="$1"
    local location="$2"

    openssl req -new -newkey rsa:4096 -nodes \
        -keyout ${location}/${username}-k8s.key \
        -out ${location}/${username}-k8s.csr \
        -subj "/CN=${username}/O=openstorage"

    cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${username}-access
spec:
  request: $(cat ${location}/${username}-k8s.csr | base64 | tr -d '\n')
  usages:
  - client auth
EOF
    kubectl certificate approve ${username}-access
    kubectl get csr ${username}-access \
        -o jsonpath='{.status.certificate}' | base64 --decode > ${location}/${username}-kubeconfig.crt
}

# Creates a new Kubernetes user only able to access their namespace with the
# same name. The kubeconfig for this user must be passed in.
function osd::createUserKubeconfig() {
    local user="$1"
    local location="$2"
    local kubeconfig="${location}/${user}-kubeconfig.conf"

    osd::createUser "$user" "$location"

    kind export kubeconfig --kubeconfig=${kubeconfig} --name ${KIND_CLUSTER}
    kubectl config set-credentials \
        ${user} \
        --client-certificate=${location}/${user}-kubeconfig.crt \
        --client-key=${location}/${user}-k8s.key \
        --embed-certs \
        --kubeconfig=${kubeconfig}
    kubectl create namespace ${user}
    kubectl --kubeconfig=${kubeconfig} config set-context ${user} \
        --cluster=kind-${KIND_CLUSTER} \
        --user=${user} \
        --namespace=${user}
    kubectl --kubeconfig=${kubeconfig} config use-context ${user}
    kubectl create rolebinding ${user}-admin --namespace=${user} --clusterrole=admin --user=${user}
}

# Delete an object in Kubernetes and wait until fully removed
function osd::kubeDeleteObjectAndWait() {
    local secs="$1"
    local kubeargs="$2"
    local object="$3"
    local name="$4"

    kubectl ${kubeargs} delete ${object} ${name}

    timeout $secs sh -c "while kubectl ${kubeargs} get ${object} ${name} > /dev/null 2>&1; do sleep 1; done "
}
