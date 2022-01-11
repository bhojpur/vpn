module github.com/bhojpur/vpn

go 1.17

require (
	github.com/lib/pq v1.10.0
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v1.1.3
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.27.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v1.5.2
)

require (
	cloud.google.com/go v0.56.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96 // indirect
	github.com/go-logr/logr v0.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.1 // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
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
