module github.com/libopenstorage/openstorage

require (
	bazil.org/fuse v0.0.0-20160317181031-37bfa8be9291
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/armon/go-metrics v0.0.0-20160521002338-fbf75676ee9c // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/cenkalti/backoff v0.0.0-20170329104900-5d150e7eec02
	github.com/cloudfoundry/gosigar v0.0.0-20150402170747-3ed7c74352da // indirect
	github.com/codegangsta/cli v0.0.0-20160326223947-bc465becccd1
	github.com/container-storage-interface/spec v1.1.0
	github.com/coreos/etcd v3.1.0-rc.1+incompatible // indirect
	github.com/coreos/go-oidc v0.0.0-20181101194249-66476e026701
	github.com/coreos/go-semver v0.2.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20160527140244-4484981625c1 // indirect
	github.com/coreos/pkg v0.0.0-20160530111557-7f080b6c11ac // indirect
	github.com/dgrijalva/jwt-go v0.0.0-20180719211823-0b96aaa70776
	github.com/docker/docker v0.0.0-20160331233925-4a7bd7eaef00
	github.com/docker/go-connections v0.3.0 // indirect
	github.com/docker/go-units v0.3.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/dustin/go-humanize v0.0.0-20151125214831-8929fe90cee4
	github.com/elazarl/goproxy v0.0.0-20190911111923-ecfe977594f1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gobuffalo/packr v1.11.0
	github.com/gogo/protobuf v0.0.0-20171018111913-117892bf1866 // indirect
	github.com/golang/mock v1.1.1
	github.com/golang/protobuf v1.2.0
	github.com/google/btree v0.0.0-20161217183710-316fb6d3f031 // indirect
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.0.0-20180218235700-15cf44e552f9 // indirect
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f // indirect
	github.com/gorilla/mux v0.0.0-20160317213430-0eeaf8392f5b
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v0.0.0-20181112102510-3304cc886352
	github.com/grpc-ecosystem/grpc-gateway v1.5.0
	github.com/hashicorp/consul v0.0.0-20160401010739-a440433ac832 // indirect
	github.com/hashicorp/go-cleanhttp v0.0.0-20160217214820-875fb671b3dd // indirect
	github.com/hashicorp/go-msgpack v0.0.0-20150518234257-fa3f63826f7c // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/logutils v0.0.0-20150609070431-0dc08b1671f3 // indirect
	github.com/hashicorp/memberlist v0.0.0-20160526233940-7c7d6bae440f // indirect
	github.com/hashicorp/serf v0.0.0-20160331225936-979180d19cb3 // indirect
	github.com/imdario/mergo v0.0.0-20181107191138-ca3dcc1022ba // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmcvetta/randutil v0.0.0-20150817122601-2bb1b664bcff // indirect
	github.com/json-iterator/go v0.0.0-20190114155330-f64ce68b6eea // indirect
	github.com/kubernetes-csi/csi-test v2.0.0+incompatible
	github.com/kubernetes-incubator/external-storage v0.0.0-00010101000000-000000000000 // indirect
	github.com/libopenstorage/gossip v0.0.0-20190507031959-c26073a01952
	github.com/libopenstorage/secrets v0.0.0-20190903232812-7dafdc1075d8
	github.com/libopenstorage/stork v0.0.0-20190115210441-3ae7cc09050f // indirect
	github.com/libopenstorage/systemutils v0.0.0-20160208220149-44ac83be3ce1
	github.com/miekg/dns v0.0.0-20160512064316-48ab6605c66a // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mohae/deepcopy v0.0.0-20161019065048-a40e8fd6a885
	github.com/onsi/ginkgo v0.0.0-20171221013426-6c46eb8334b3
	github.com/onsi/gomega v1.3.0
	github.com/opencontainers/runc v0.0.0-20160331090202-89ab7f2ccc1e // indirect
	github.com/operator-framework/operator-sdk v0.0.7 // indirect
	github.com/pborman/uuid v0.0.0-20160216163710-c55201b03606
	github.com/peterbourgon/diskv v0.0.0-20180312054125-0646ccaebea1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/portworx/kvdb v0.0.0-20190206214116-2083b561383e
	github.com/portworx/sched-ops v0.0.0-20190115193420-d64ebc777d0c
	github.com/portworx/talisman v0.0.0-20190115023107-286f67c146ff // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rs/cors v0.0.0-20190116175910-76f58f330d76
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.0-20160331143210-b0d571e7d5f7
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.2.2
	github.com/ugorji/go v0.0.0-20160328060740-a396ed22fc04 // indirect
	github.com/urfave/negroni v0.0.0-20181201104632-7183f09c600e
	github.com/vbatts/tar-split v0.0.0-20160330203851-226f7c74905f // indirect
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9 // indirect
	golang.org/x/net v0.0.0-20181106065722-10aee1819953
	golang.org/x/oauth2 v0.0.0-20181203162652-d668ce993890 // indirect
	golang.org/x/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/time v0.0.0-20170927054726-6dc17368e09b // indirect
	google.golang.org/appengine v1.3.0 // indirect
	google.golang.org/genproto v0.0.0-20181202183823-bd91e49a0898
	google.golang.org/grpc v1.17.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/jmcvetta/napping.v3 v3.0.5
	gopkg.in/square/go-jose.v2 v2.2.1 // indirect
	gopkg.in/vmihailenco/msgpack.v2 v2.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.1
	k8s.io/api v0.0.0-20180628040859-072894a440bd // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20180628053655-3de98c57bc05 // indirect
	k8s.io/apimachinery v0.0.0-20180621070125-103fd098999d // indirect
	k8s.io/client-go v8.0.0+incompatible // indirect
)

replace github.com/kubernetes-incubator/external-storage => github.com/libopenstorage/external-storage v5.2.0-openstorage+incompatible

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5

replace github.com/docker/distribution/digest => github.com/opencontainers/go-digest v0.0.0-20190306001800-ac19fd6e74
