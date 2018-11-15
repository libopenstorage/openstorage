#!/bin/sh
# Simple script to add generated files from api.proto into their own commit
git add api/api.pb.g* \
	api/server/sdk/api/api.swagger.json \
	pkg/jsonpb/testing/testing.pb.go
