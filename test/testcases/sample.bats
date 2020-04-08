load ../lib/detik
load ../'node_modules/bats-assert/load'

DETIK_CLIENT_NAME="kubectl -n kube-system"

@test "This is a sample" {
	run kubectl get nodes
	assert_success
}

@test "Verify openstorage running" {
	verify "there is 1 pod named 'openstorage'"
}
