set -x 

#!/bin/bash
fail() {
	echo "$1"
	exit 1
}

registerFuncs=$(cat api/api.pb.go | grep 'func Register.*' | wc -l)
registered=$(cat api/server/sdk/rest_gateway.go | grep 'Register.*Handler,' | wc -l)

if [ "$registerFuncs" != "$registered" ] ; then
	fail "Make sure all REST handlers are registered in api/server/sdk/rest_gateway.go. Check api/api.pb.go for missing handlers (i.e. RegisterOpenStorageWatchHandler)."
else
	echo "All REST handlers are registered in api/server/sdk/rest_gateway.go."
fi
