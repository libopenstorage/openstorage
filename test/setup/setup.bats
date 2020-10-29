load ../vendor/k8s
load ../lib/osd
load ../node_modules/bats-assert/load
load ../node_modules/bats-support/load

ASSETS="setup/assets"

@test "Setup kind cluster ${KIND_CLUSTER}" {
    local name=${KIND_CLUSTER}
    if kind get clusters | grep ${KIND_CLUSTER} > /dev/null 2>&1 ; then
        skip "Cluster already up and running"
    fi

    run kind create cluster \
        --name ${name} \
        --config ${ASSETS}/kind.yaml \
        --image ${KIND_IMAGE}
    assert_success

    run kubectl apply -f ${ASSETS}/noauth
    assert_success

    run kubectl apply -f ${ASSETS}/auth
    assert_success

    run kubectl apply -f ${ASSETS}/multitenant
    assert_success

    run kubectl create namespace openstorage
    assert_success

    run kubectl -n openstorage create secret \
        generic k8s-user --from-literal=auth-token=${K8S_TOKEN}
    assert_success

    run kubectl -n openstorage create secret \
        generic admin-user --from-literal=auth-token=${ADMIN_TOKEN}
    assert_success
}

@test "Wait for Kind cluster to be ready" {
    # Wait until all the pod statuses are not in Running state
    run timeout 120 bash -c \
      "while kubectl get pods -A -o json | jq -r '.items[].status.phase == \"Running\"' | grep false > /dev/null 2>&1 ; do sleep 1; done"
    assert_success

    run timeout 120 bash -c "while kubectl get nodes | grep NotReady > /dev/null 2>&1; do sleep 1; done"
    assert_success
}

@test "Install openstorage in ${KIND_CLUSTER}" {
    run kind load docker-image quay.io/openstorage/osd:latest --name ${KIND_CLUSTER} || exit 1
    assert_success

    # Start OSD
    kubectl delete -f ${ASSETS}/osd-csi.yaml > /dev/null 2>&1 || true

    # Deploy
    run kubectl apply -f ${ASSETS}/osd-csi.yaml
    assert_success

    # Tell DETIK what command to use to verify
    DETIK_CLIENT_NAME="kubectl -n kube-system"

    # Wait for openstorage to come up
    run try "at most 120 times every 1s to get pods named '^openstorage' and verify that 'status' is 'running'"
    assert_success

}

@test "Verify SDK GW is accessible" {
    timeout 60 sh -c "until curl --silent -H \"Authorization:bearer $ADMIN_TOKEN\" -X GET -d {} http://$(osd::getSdkRestGWEndpoint)/v1/clusters/inspectcurrent | grep STATUS_OK; do sleep 1; done"
}
