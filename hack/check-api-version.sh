#!/bin/bash

fail() {
	echo "$1"
	echo ""
	echo "Please update the SdkVersion and SDK_CHANGELOG.md"
	exit 1
}

# Check if the api.proto was changed
# Don't check versions if only the comments have been updated.
if ! git diff master..HEAD api/api.proto | grep -v "^\+\+\+../api/api.proto$" | grep "^\+" | grep -v "^\+.*//" > /dev/null 2>&1 ; then
	exit 0
fi

currentver=$(go run tools/sdkver/sdkver.go)
prevver=$(git show master:api/server/sdk/api/api.swagger.json | jq -r '.info.version')

if [ "$currentver" = "$prevver" ] ; then
	fail "SdkVersion $currentver matches previous version and has not been updated"
fi

clver=$(egrep -o "([0-9]{1,}\.)+[0-9]{1,}" SDK_CHANGELOG.md  | head -1)
if [ "$currentver" != "$clver" ] ; then
	fail "SdkVersion of $currentver in api.proto does not match latest SDK_CHANGELOG.md of $clver"
fi
