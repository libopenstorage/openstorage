#!/bin/sh

BART=./node_modules/bats/bin/bats
export CLUSTER_CONTROL_PLANE_CONTAINER=lpabon-kind-csi-control-plane

${BART} --tap testcases

