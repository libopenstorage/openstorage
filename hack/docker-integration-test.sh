#!/bin/bash
set -x

cleanup() {
	sudo pkill -15 -f $GOPATH/bin/osd
}

assert_success() {
	if [ $? -ne 0 ];
	then
		cleanup
		exit 1
	fi
}

get_volume_id() {
	volume_id=$(curl -X POST "http://localhost:9116/v1/volumes/filters" -H "accept: application/json" -H "Content-Type:application/json" -H "Authorization:bearer $token" -d "{ \"name\": \"$1\"}" | jq .[] | jq .[0] -r)
}

assert_attached(){
	# Get Vol ID
	get_volume_id $volume_name
	attach_path=$(curl -X GET "http://localhost:9116/v1/volumes/inspect/$volume_id" -H "accept:application/json" -H "Authorization:bearer $token" | jq  ."volume" | jq ."attach_path" | jq .[0] -r)

	# ATTACH_PATH is null when the volume unmounted. If it is non-null,
	# we know that the volume is mounted.
	#
	# Here we can assert that the volume is mounted
	if [ $1 == "true" ];
	then
		echo "Asserting volume attached: $attach_path"
		if [ $attach_path != "null" ];
		then
			return 0
		else
			cleanup
			exit 1
		fi
	# Here we can assert that the volume is unmounted
	else
		echo "Asserting volume detached: $attach_path"
		if [ $attach_path == "null" ];
		then
			return 0
		else
			cleanup
			exit 1
		fi
	fi
}

# Generate shared secret
make install
token=$($GOPATH/bin/osd-token-generator \
  --auth-config=hack/sdk-auth-sample.yml \
  --shared-secret=testsecret)

# Start OSD
sudo -E $GOPATH/bin/osd \
	-d \
	--driver=name=fake \
	--sdkport 9106 \
	--jwt-shared-secret=testsecret \
	--sdkrestport 9116 &
jobs -l
sleep 3

# Test & assert
volume_name=$(sudo docker volume create -d fake -o size=1234 -o token=$token)
assert_success
sudo docker volume inspect token=$token,name=${volume_name}
assert_success

# Get Vol ID
get_volume_id $volume_name

# Check if volume is unmounted
assert_attached false

# Run app with our  volume
app_name="APP_TEST_$volume_name"
sudo docker run -d --name $app_name --volume-driver fake -v size=12345,token=$token,name=${volume_name}_new:/app nginx:latest
assert_success
assert_attached true

# Unmount, remove
sudo docker stop $app_name
assert_success
assert_attached false

# Cleanup
echo "Docker integration tests passed, cleaning up!"
cleanup
