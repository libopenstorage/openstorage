#!/bin/bash
KIND_IMAGE=kindest/node:v1.17.0
set -x
cd $GOPATH/src/github.com/libopenstorage/openstorage

# Setup KinD cluster
if [ $DELETE_CLUSTER ]; then
	kind delete cluster --name kind-csi 
fi
kind create cluster --name kind-csi --config hack/kind.yaml --image $KIND_IMAGE

# Build OSD
make docker-build-osd

# Load local OSD image into KinD
kind load docker-image quay.io/openstorage/osd:latest --name kind-csi
export KUBECONFIG=$(kind get kubeconfig-path --name kind-csi)

# Start OSD
kubectl delete -f hack/osd-csi.yaml
kubectl apply -f hack/osd-csi.yaml

set +x
echo "Done, check your cluster to see if the CSI sidecars are ready"
echo "export KUBECONFIG=$(kind get kubeconfig-path --name kind-csi); alias kk=kubectl;alias kn='kubectl -n kube-system'"
