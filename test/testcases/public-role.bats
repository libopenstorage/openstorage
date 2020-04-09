load ../vendor/k8s
load ../lib/osd
load ../node_modules/bats-assert/load
load ../node_modules/bats-support/load

ASSETS=testcases/assets

@test "Verify a user can create a public volume" {
    local user="user$$"
    local kubeconfig="${BATS_TMPDIR}/${user}-kubeconfig.conf"
    run osd::createUserKubeconfig "${user}" "$BATS_TMPDIR"
    assert_success

    # Tell DETIK what command to use to verify
    DETIK_CLIENT_NAME="kubectl -n ${user}"

    # create pvc
    run kubectl --kubeconfig=${kubeconfig} create -f ${ASSETS}/pvc-noncsi.yml
    assert_success

    # assert it is there
    run try "at most 100 times every 1s to get pvc named 'mypvctest' and verify that 'status' is 'bound'"

    # cleanup
    run kubectl --kubeconfig=${kubeconfig} delete -f ${ASSETS}/pvc-noncsi.yml
}
