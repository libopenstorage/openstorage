
TMPDIR="${BATS_TMPDIR:-/tmp}"
KIND_CLUSTER="${KIND_CLUSTER:-lpabon-kind-csi}"


function osd::clusterip() {
    docker inspect ${CLUSTER_CONTROL_PLANE_CONTAINER} | jq -r '.[].NetworkSettings.Networks.bridge.IPAddress'
}

function osd::getSdkRestGWEndpoint() {
    local clusterip=$(osd::clusterip)
    local nodeport=$(kubectl -n kube-system get svc portworx-api -o json | jq '.spec.ports[2].nodePort')
    echo ${clusterip}:${nodeport}
}

function osd::getSdkEndpoint() {
    local clusterip=$(osd::clusterip)
    local nodeport=$(kubectl -n kube-system get svc portworx-api -o json | jq '.spec.ports[1].nodePort')
    echo ${clusterip}:${nodeport}
}

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
