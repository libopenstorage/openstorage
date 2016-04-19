#!/bin/bash

set -Ee

DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${DIR}"

if [[ "${TRAVIS_BRANCH}" != "master" ]]; then
  echo "TRAVIS_BRANCH is ${TRAVIS_BRANCH}, which is not master, will not docker push openstorage/osd" >&2
  exit 0
fi

if [[ -z "${DOCKER_EMAIL}" ]]; then
  echo "error: DOCKER_EMAIL not set" >&2
  exit 1
fi

if [[ -z "${DOCKER_USER}" ]]; then
  echo "error: DOCKER_USER not set" >&2
  exit 1
fi

if [[ -z "${DOCKER_PASS}" ]]; then
  echo "error: DOCKER_PASS not set" >&2
  exit 1
fi

make docker-build-osd
docker login -e "${DOCKER_EMAIL}" -u "${DOCKER_USER}" -p "${DOCKER_PASS}"
docker push openstorage/osd
