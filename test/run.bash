#!/bin/sh

BART=./node_modules/bats/bin/bats

export KIND_CLUSTER=lpabon-kind-csi
export CLUSTER_CONTROL_PLANE_CONTAINER=${KIND_CLUSTER}-control-plane
export TMPDIR=/tmp/bats-test-$$
mkdir -p ${TMPDIR}

${BART} --tap testcases

