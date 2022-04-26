module github.com/bhojpur/vpn

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/benbjohnson/clock v1.3.0
	github.com/c-robinson/iplib v1.0.3
	github.com/hashicorp/golang-lru v0.5.4
	github.com/ipfs/go-log v1.0.5
	github.com/ipfs/go-log/v2 v2.5.1
	github.com/labstack/echo/v4 v4.7.2
	github.com/libp2p/go-libp2p v0.19.0
	github.com/libp2p/go-libp2p-connmgr v0.4.0
	github.com/libp2p/go-libp2p-core v0.15.1
	github.com/libp2p/go-libp2p-discovery v0.7.0
	github.com/libp2p/go-libp2p-kad-dht v0.15.0
	github.com/libp2p/go-libp2p-mplex v0.7.0
	github.com/libp2p/go-libp2p-pubsub v0.6.1
	github.com/libp2p/go-libp2p-resource-manager v0.3.0
	github.com/libp2p/go-libp2p-yamux v0.9.1
	github.com/miekg/dns v1.1.48
	github.com/multiformats/go-multiaddr v0.5.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.19.0
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/pterm/pterm v0.12.41
	github.com/sirupsen/logrus v1.8.1
	github.com/songgao/packets v0.0.0-20160404182456-549a10cd4091
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/spf13/cobra v1.4.0
	github.com/urfave/cli v1.22.7
	github.com/vishvananda/netlink v1.1.0
	github.com/xlzd/gotp v0.0.0-20220110052318-fab697c03c2c
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.23.6
	k8s.io/client-go v1.5.2
)

require (
	cloud.google.com/go/compute v1.6.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/docker/spdystream v0.1.0 // indirect
	github.com/fsnotify/fsnotify v1.5.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/ipfs/go-cid v0.2.0 // indirect
	github.com/ipld/go-ipld-prime v0.16.0 // indirect
	github.com/klauspost/compress v1.15.2 // indirect
	github.com/libp2p/go-libp2p-asn-util v0.2.0 // indirect
	github.com/libp2p/go-reuseport v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/multiformats/go-multistream v0.3.1 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74 // indirect
	golang.org/x/term v0.0.0-20220411215600-e5f449aeb171 // indirect
	golang.org/x/time v0.0.0-20220411224347-583f2d630306 // indirect
	google.golang.org/genproto v0.0.0-20220422154200-b37d22cd5731 // indirect
	k8s.io/api v0.23.6 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace k8s.io/api => k8s.io/api v0.20.4

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.4

replace k8s.io/apimachinery => k8s.io/apimachinery v0.20.4

replace k8s.io/apiserver => k8s.io/apiserver v0.20.4

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.4

replace k8s.io/client-go => k8s.io/client-go v0.20.4

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.4

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.4

replace k8s.io/code-generator => k8s.io/code-generator v0.20.4

replace k8s.io/component-base => k8s.io/component-base v0.20.4

replace k8s.io/cri-api => k8s.io/cri-api v0.20.4

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.4

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.4

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.4

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.4

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.4

replace k8s.io/kubelet => k8s.io/kubelet v0.20.4

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.4

replace k8s.io/metrics => k8s.io/metrics v0.20.4

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.4

replace k8s.io/component-helpers => k8s.io/component-helpers v0.20.4

replace k8s.io/controller-manager => k8s.io/controller-manager v0.20.4

replace k8s.io/kubectl => k8s.io/kubectl v0.20.4

replace k8s.io/mount-utils => k8s.io/mount-utils v0.20.4
