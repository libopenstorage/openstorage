#!/bin/bash
set -x
echo "$DOCKERPASS" | docker login -u "$DOCKERUSER" --password-stdin
git-validation -run DCO,short-subject
go fmt $(go list ./... | grep -v vendor) | grep -v "api.pb.go" | wc -l | grep "^0";
make docker-proto
git diff $(find . -name "*.pb.*go" -o -name "api.swagger.json" | grep -v vendor) | wc -l | grep "^0"
hack/check-api-version.sh
git grep -rw GPL vendor | grep LICENSE | egrep -v "yaml.v2" | wc -l | grep "^0"
bash hack/docker-integration-test.sh
make install verify
if [ "${TRAVIS_PULL_REQUEST}" == "false" ]; then echo "$DOCKERPASS" | docker login -u "$DOCKERUSER" --password-stdin; make push-mock-sdk-server; fi