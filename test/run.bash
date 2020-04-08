#!/bin/sh

BART=./node_modules/bats/bin/bats

# generate 10y tokens
export ADMIN_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN1cHBvcnQtYWRtaW5AbXljb21wYW55LmNvbSIsImV4cCI6MTkwMTczMDQwNywiZ3JvdXBzIjpbIioiXSwiaWF0IjoxNTg2MzcwNDA3LCJpc3MiOiJvcGVuc3RvcmFnZS5pbyIsIm5hbWUiOiJBZG1pbiIsInJvbGVzIjpbInN5c3RlbS5hZG1pbiJdLCJzdWIiOiJzdXBwb3J0LWFkbWluQG15Y29tcGFueS5jb20ifQ.RR0hduw2x4aQPLUFzwXRMp3g0Qg1Uq-gGkIY-vCMxRE
export K8S_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN1cHBvcnRAbXljb21wYW55LmNvbSIsImV4cCI6MTkwMTczMDQ3MywiZ3JvdXBzIjpbIm15Y29tcGFueSIsImVuZ2luZWVyaW5nIiwiZGV2b3BzIl0sImlhdCI6MTU4NjM3MDQ3MywiaXNzIjoib3BlbnN0b3JhZ2UuaW8iLCJuYW1lIjoiS3ViZXJuZXRlcyIsInJvbGVzIjpbInN5c3RlbS51c2VyIl0sInN1YiI6InN1cHBvcnRAbXljb21wYW55LmNvbSJ9.2EnoEAR2qrTTxjxcH3k5w_E24l4p5DU7jOWF7ke0aJ4
export TENANT1_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN1cHBvcnRAdGVuYW50LW9uZS5jb20iLCJleHAiOjE5MDE3MzA1MDUsImdyb3VwcyI6WyJ0ZW5hbnQtb25lIl0sImlhdCI6MTU4NjM3MDUwNSwiaXNzIjoib3BlbnN0b3JhZ2UuaW8iLCJuYW1lIjoiVGVuYW50IE9uZSIsInJvbGVzIjpbInN5c3RlbS51c2VyIl0sInN1YiI6InN1cHBvcnRAdGVuYW50LW9uZS5jb20ifQ.56ruILoD_r-RpE_r9317nWq8gZ7PbjnMY5JMzQrPuhI
export TENANT2_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN1cHBvcnRAdGVuYW50LXR3by5jb20iLCJleHAiOjE5MDE3MzA1NDksImdyb3VwcyI6WyJ0ZW5hbnQtdHdvIl0sImlhdCI6MTU4NjM3MDU0OSwiaXNzIjoib3BlbnN0b3JhZ2UuaW8iLCJuYW1lIjoiVGVuYW50IFR3byIsInJvbGVzIjpbInN5c3RlbS51c2VyIl0sInN1YiI6InN1cHBvcnRAdGVuYW50LXR3by5jb20ifQ.6t3DiToB5ttTKZ9IuSoM4XTKKltpBq84kz7HseehjFc

export KIND_CLUSTER=lpabon-kind-csi
export CLUSTER_CONTROL_PLANE_CONTAINER=${KIND_CLUSTER}-control-plane
export TMPDIR=/tmp/bats-test-$$
mkdir -p ${TMPDIR}

${BART} --tap testcases

