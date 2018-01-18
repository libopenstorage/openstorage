#!/usr/bin/env bash
protoc -I . config.proto --go_out=plugins=grpc:.
