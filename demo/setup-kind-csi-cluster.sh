#!/bin/bash
KIND_IMAGE=kindest/node:v1.17.0
set -x
cd $GOPATH/src/github.com/libopenstorage/openstorage

# 1. Build openstorage
make docker-build-osd

# 2. Setup KinD cluster
if [ $DELETE_CLUSTER ]; then
	kind delete cluster --name openstorage-test-cluster
fi
kind create cluster --name openstorage-test-cluster --config hack/kind.yaml --image $KIND_IMAGE
export KUBECONFIG=$(kind get kubeconfig-path --name openstorage-test-cluster)

# 3. Load local openstorage image into KinD
kind load docker-image quay.io/openstorage/osd:latest --name openstorage-test-cluster

# 4. Start OSD
kubectl delete -f openstorage-setup/
kubectl apply -f openstorage-setup/

set +x
echo "Done, check your cluster to see if the CSI sidecars are ready"
echo "export KUBECONFIG=$(kind get kubeconfig-path --name openstorage-test-cluster)"
