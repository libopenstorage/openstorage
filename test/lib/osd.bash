
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

