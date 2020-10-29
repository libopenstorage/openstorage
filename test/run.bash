#!/bin/bash

# Load config
source ./config

# Check dependencies
dependecies="kind kubectl jq curl"
for d in $dependecies ; do
    if [[ -z "$(type -t $d)" ]] ; then
        echo "Missing $d" >&2
        exit 1
    fi
done

# Location of bart
BART=./node_modules/bats/bin/bats

# Set env DEBUG=1 to show output of osd::echo and osd::by
export TMPDIR=/tmp/bats-test-$$
mkdir -p ${TMPDIR} && \
    ${BART} setup testcases && \
    rm -rf ${TMPDIR}
