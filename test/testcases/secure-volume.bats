load ../vendor/k8s
load ../lib/osd
load ../node_modules/bats-assert/load
load ../node_modules/bats-support/load

ASSETS=testcases/assets

@test "Verify user can create a pvc with authentication" {
    local pvcname="pvc-auth"
    local user="user$$"
    local kubeconfig="${BATS_TMPDIR}/${user}-kubeconfig.conf"

    DETIK_CLIENT_NAME="kubectl -n ${user}"
    run osd::createUserKubeconfig "${user}" "$BATS_TMPDIR"
    assert_success

    storageclasses="intree-auth csi-auth"
    for sc in $storageclasses ; do
        sed -e "s#%%PVCNAME%%#${pvcname}#" \
        -e "s#%%STORAGECLASS%%#${sc}#" \
        ${ASSETS}/pvc.yml.tmpl | kubectl --kubeconfig=${kubeconfig} create -f -

        # assert it is there
        run try "at most 120 times every 1s to get pvc named '^${pvcname}' and verify that 'status' is 'bound'"
        assert_success

        # assert that the owner is the tenant. The 'sub' for the kubernetes
        # token is: support@mycompany.com, so this *must* be the owner of the volume.
        # Since we have created only one, there must be exactly only 1 volume owned
        # by this account
        nvols=$(curl -s -X POST \
            "http://$(osd::getSdkRestGWEndpoint)/v1/volumes/inspectwithfilters" \
            -H "accept: application/json" \
            -H "Content-Type: application/json" \
            -H "Authorization: bearer $K8S_TOKEN" \
            -d "{\"ownership\":{\"owner\":\"support@mycompany.com\"}}" | jq '.volumes | length')
        [[ $nvols -eq 1 ]]

        # cleanup
        kubectl --kubeconfig=${kubeconfig} delete pvc ${pvcname}
    done
}

@test "Verify multitenancy by having user create volume with their token" {
    local pvcname="pvc-auth"
    local user="tenant-1-$$"
    local kubeconfig="${BATS_TMPDIR}/${user}-kubeconfig.conf"

    DETIK_CLIENT_NAME="kubectl -n ${user}"
    run osd::createUserKubeconfig "${user}" "$BATS_TMPDIR"
    assert_success

    # Insert token as admin
    run kubectl -n ${user} create secret \
        generic k8s-user --from-literal=auth-token=${TENANT1_TOKEN}
    assert_success

    storageclasses="intree-multitenant csi-multitenant"
    for sc in $storageclasses ; do
        sed -e "s#%%PVCNAME%%#${pvcname}#" \
        -e "s#%%STORAGECLASS%%#${sc}#" \
        ${ASSETS}/pvc.yml.tmpl | kubectl --kubeconfig=${kubeconfig} create -f -

        # assert it is there
        run try "at most 120 times every 1s to get pvc named '^${pvcname}' and verify that 'status' is 'bound'"
        assert_success

        # assert that the owner is the tenant. The 'sub' for the tenant 1
        # token is: support@tenant-one.com, so this *must* be the owner of the volume.
        # Since we have created only one, there must be exactly only 1 volume owned
        # by this account
        nvols=$(curl -s -X POST \
            "http://$(osd::getSdkRestGWEndpoint)/v1/volumes/inspectwithfilters" \
            -H "accept: application/json" \
            -H "Content-Type: application/json" \
            -H "Authorization: bearer $TENANT1_TOKEN" \
            -d "{\"ownership\":{\"owner\":\"support@tenant-one.com\"}}" | jq '.volumes | length')
        [[ $nvols -eq 1 ]]

        # cleanup
        kubectl --kubeconfig=${kubeconfig} delete pvc ${pvcname}
    done
}