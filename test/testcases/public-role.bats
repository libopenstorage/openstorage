load ../lib/detik
load ../node_modules/bats-assert/load

DETIK_CLIENT_NAME="kubectl -n kube-system"

@test "Verify openstorage running" {
	verify "there is 1 pod named 'openstorage'"
}
