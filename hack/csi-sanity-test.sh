#!/bin/bash
set -x

cleanup() {
	sudo pkill -15 -f $GOPATH/bin/osd
	go mod tidy
	go mod vendor
}

assert_success() {
	if [ $? -ne 0 ];
	then
		cleanup
		exit 1
	fi
}

# Install osd binary
make install
assert_success

# Start OSD
sudo -E $GOPATH/bin/osd \
	-d \
	--driver=name=fake \
	--sdkport 9106 \
	--sdkrestport 9116 &
jobs -l

# Run CSI Test
go get -u github.com/kubernetes-csi/csi-test/...
sudo $GOPATH/bin/csi-sanity --csi.endpoint=/var/lib/osd/driver/fake-csi.sock
assert_success

# Cleanup
echo "CSI sanity tests passed, cleaning up!"
cleanup
