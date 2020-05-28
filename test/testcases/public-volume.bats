load ../vendor/k8s
load ../lib/osd
load ../node_modules/bats-assert/load
load ../node_modules/bats-support/load

ASSETS=testcases/assets

@test "Verify user can create a pvc without authentication" {
    local pvcname="pvc-noauth"
    local user="user$$"
    local kubeconfig="${BATS_TMPDIR}/${user}-kubeconfig.conf"

    DETIK_CLIENT_NAME="kubectl -n ${user}"
    run osd::createUserKubeconfig "${user}" "$BATS_TMPDIR"
    assert_success

    storageclasses="intree-noauth csi-noauth"
    for sc in $storageclasses ; do
        sed -e "s#%%PVCNAME%%#${pvcname}#" \
        -e "s#%%STORAGECLASS%%#${sc}#" \
        ${ASSETS}/pvc.yml.tmpl | kubectl --kubeconfig=${kubeconfig} create -f -

        # assert it is there
        run try "at most 120 times every 1s to get pvc named '^${pvcname}' and verify that 'status' is 'bound'"
        assert_success

        # cleanup
        run osd::kubeDeleteObjectAndWait 120 "--kubeconfig=${kubeconfig}" "pvc" "${pvcname}"
        assert_success
    done
}

