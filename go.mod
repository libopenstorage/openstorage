module github.com/libopenstorage/openstorage

go 1.15

require (
	bazil.org/fuse v0.0.0-20160317181031-37bfa8be9291
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cloudfoundry/gosigar v0.0.0-20150402170747-3ed7c74352da // indirect
	github.com/codegangsta/cli v1.13.1-0.20160326223947-bc465becccd1
	github.com/container-storage-interface/spec v1.5.0
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/dgrijalva/jwt-go v3.2.1-0.20180719211823-0b96aaa70776+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v1.1.0 // indirect
	github.com/gobuffalo/packr v1.11.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/consul/api v1.8.1 // indirect
	github.com/hashicorp/go-hclog v0.12.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/kubernetes-csi/csi-test v2.2.0+incompatible
	github.com/kubernetes-csi/csi-test/v4 v4.2.0 // indirect
	github.com/libopenstorage/gossip v0.0.0-20200808224301-d5287c7c8b24
	github.com/libopenstorage/secrets v0.0.0-20200207034622-cdb443738c67
	github.com/libopenstorage/systemutils v0.0.0-20160208220149-44ac83be3ce1
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/miekg/dns v1.1.35 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/symlink v0.1.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/portworx/kvdb v0.0.0-20200723230726-2734b7f40194
	github.com/portworx/sched-ops v1.20.4-rc1
	github.com/prometheus/client_golang v1.9.0
	github.com/robertkrimen/otto v0.0.0-20210614181706-373ff5438452 // indirect
	github.com/rs/cors v1.6.1-0.20190116175910-76f58f330d76
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/negroni v1.0.1-0.20181201104632-7183f09c600e
	github.com/vbatts/tar-split v0.9.14-0.20160330203851-226f7c74905f // indirect
	go.pedge.io/pb v0.0.0-20171203174523-dbc791b8a69c // indirect
	go.pedge.io/proto v0.0.0-00010101000000-000000000000
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/net v0.0.0-20210908191846-a5e095526f91
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210909211513-a8c4777a87af
	google.golang.org/grpc v1.40.0
	gopkg.in/freddierice/go-losetup.v1 v1.0.0-20170407175016-fc9adea44124
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.20.0 // indirect
)

replace (
	github.com/container-storage-interface/spec => github.com/container-storage-interface/spec v1.5.0
	github.com/docker/docker => github.com/moby/moby v20.10.3-0.20210324213045-797b974cb90e+incompatible
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/kubernetes-incubator/external-storage => github.com/libopenstorage/external-storage v0.20.4-openstorage-rc3
	github.com/satori/go.uuid => github.com/satori/go.uuid v0.0.0-20160324112244-f9ab0dce87d8

	go.pedge.io/proto => go.pedge.io/proto v0.0.0-20170422232847-c5da4db108f6
	google.golang.org/grpc => google.golang.org/grpc v1.29.1

	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.4
	k8s.io/apiserver => k8s.io/apiserver v0.20.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.4
	k8s.io/code-generator => k8s.io/code-generator v0.20.4
	k8s.io/component-base => k8s.io/component-base v0.20.4
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.4
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.4
	k8s.io/cri-api => k8s.io/cri-api v0.20.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.4
	k8s.io/kubectl => k8s.io/kubectl v0.20.4
	k8s.io/kubelet => k8s.io/kubelet v0.20.4
	k8s.io/kubernetes => k8s.io/kubernetes v1.20.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.4
	k8s.io/metrics => k8s.io/metrics v0.20.4
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.4
)
