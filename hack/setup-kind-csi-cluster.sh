#!/bin/bash
set -x
cd $GOPATH/src/github.com/libopenstorage/openstorage
if [ $DELETE_CLUSTER ]; then
	kind delete cluster --name kind-csi 
fi
kind create cluster --name kind-csi --config hack/kind.yaml --image kindest/node:v1.17.0
docker cp $GOPATH/bin/osd kind-csi-worker:/osdkind
docker exec -ti kind-csi-worker bash -c "chmod 777 /osdkind"
docker exec -ti kind-csi-worker bash -c "mkdir -p /var/lib/kubelet/plugins/osd.openstorage.org"
docker exec -ti kind-csi-worker bash -c "pkill osdkind"
docker exec -tid kind-csi-worker bash -c "CSI_ENDPOINT=/var/lib/kubelet/plugins/osd.openstorage.org/csi.sock /osdkind -d --driver=name=fake --sdkport 9106 --sdkrestport 9116 --csidrivername osd.openstorage.org"
export KUBECONFIG=$(kind get kubeconfig-path --name kind-csi)
kubectl apply -f hack/osd-csi.yaml

set +x
echo "Done, check your cluster to see if the CSI sidecars are ready"
echo "export KUBECONFIG=$(kind get kubeconfig-path --name kind-csi); alias kk=kubectl;alias kn='kubectl -n kube-system'"
