load ../vendor/k8s
load ../'node_modules/bats-assert/load'

@test "This is a sample" {
    run kubectl get nodes
    assert_success
}

