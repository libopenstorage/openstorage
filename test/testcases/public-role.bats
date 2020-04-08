load ../lib/detik
load ../lib/osd
load ../node_modules/bats-assert/load

DETIK_CLIENT_NAME="kubectl -n kube-system"

@test "Verify openstorage running" {
	verify "there is 1 pod named 'openstorage'"
}

@test "Verify SDK GW is accessible" {
	run curl http://$(osd::getSdkRestGWEndpoint)/v1/identities/version
	assert_success
}
