module github.com/libopenstorage/openstorage

go 1.17

require (
	bazil.org/fuse v0.0.0-20160317181031-37bfa8be9291
	github.com/aws/aws-sdk-go v1.44.24
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/cloudfoundry/gosigar v1.3.3 // indirect
	github.com/codegangsta/cli v1.13.1-0.20160326223947-bc465becccd1
	github.com/container-storage-interface/spec v1.6.0
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.1-0.20180719211823-0b96aaa70776+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/gobuffalo/packr v1.11.0
	github.com/golang/mock v1.3.2-0.20190713102442-dd8d2a22370e
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.7.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20181112102510-3304cc886352
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.11.3
	github.com/hashicorp/go-rootcerts v1.0.3-0.20191216101743-c8a9a31cbd76 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/johannesboyne/gofakes3 v0.0.0-20210819161434-5c8dfcfe5310
	github.com/kubernetes-csi/csi-test v2.0.0+incompatible
	github.com/libopenstorage/gossip v0.0.0-20220309192431-44c895e0923e
	github.com/libopenstorage/secrets v0.0.0-20200207034622-cdb443738c67
	github.com/libopenstorage/systemutils v0.0.0-20160208220149-44ac83be3ce1
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.5.0
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/portworx/kvdb v0.0.0-20191122191654-7190161df3b0
	github.com/portworx/sched-ops v0.0.0-20220525225947-28b8d876887b
	github.com/prometheus/client_golang v1.7.1
	github.com/rs/cors v1.6.1-0.20190116175910-76f58f330d76
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.5.1
	github.com/urfave/negroni v1.0.1-0.20181201104632-7183f09c600e
	github.com/vbatts/tar-split v0.9.14-0.20160330203851-226f7c74905f // indirect
	go.pedge.io/env v0.0.0-20171203202719-5f5a7ded0de0
	go.pedge.io/lion v0.0.0-20190619200210-304b2f426641
	go.pedge.io/pb v0.0.0-20171203174523-dbc791b8a69c // indirect
	go.pedge.io/proto v0.0.0-20170422232847-c5da4db108f6
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220405210540-1e041c57c461
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.28.1-0.20220722155247-2a74a0e82391 // indirect
	gopkg.in/freddierice/go-losetup.v1 v1.0.0-20170407175016-fc9adea44124
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.1-beta.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1

	github.com/kubernetes-incubator/external-storage => github.com/libopenstorage/external-storage v5.1.1-0.20190919185747-9394ee8dd536+incompatible
	github.com/kubernetes-incubator/external-storage v0.0.0-00010101000000-000000000000 => github.com/libopenstorage/external-storage v5.1.1-0.20190919185747-9394ee8dd536+incompatible
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/api => k8s.io/api v0.15.11
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.15.11
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.11
	k8s.io/apiserver => k8s.io/apiserver v0.15.11
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.15.11
	k8s.io/client-go => k8s.io/client-go v0.15.11
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.15.11
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.15.11
	k8s.io/code-generator => k8s.io/code-generator v0.15.11
	k8s.io/component-base => k8s.io/component-base v0.15.11
	k8s.io/cri-api => k8s.io/cri-api v0.15.11
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.15.11
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.15.11
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.15.11
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.15.11
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.15.11
	k8s.io/kubectl => k8s.io/kubectl v0.15.11
	k8s.io/kubelet => k8s.io/kubelet v0.15.11
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.15.11
	k8s.io/metrics => k8s.io/metrics v0.15.11
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.15.11
)
