#!/bin/sh
set -x

cleanup() {
	rm .osd_integration_test_id
	sudo pkill -15 -f $GOPATH/bin/osd
}

assert_success() {
	if [ $? -ne 0 ];
	then
		cleanup
		exit 1
	fi
}

# Start OSD
make install
sudo $GOPATH/bin/osd -d --driver=name=nfs,server=127.0.0.1,path=/nfs --sdkport 9106 --sdkrestport 9116 &
jobs -l

# Test & assert
sudo docker volume create -d nfs -o size=1234 > .osd_integration_test_id
assert_success
sudo docker volume inspect `cat .osd_integration_test_id`
assert_success
sudo docker volume rm `cat .osd_integration_test_id`
assert_success

# Cleanup
cleanup
