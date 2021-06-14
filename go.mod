module github.com/appuio/openshift-machineset-egress-cidr-operator

go 1.16

require (
	github.com/openshift/api v0.0.0-20210428205234-a8389931bee7
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/openshift/machine-api-operator v0.2.1-0.20210521181620-e179bb5ce397
	k8s.io/apimachinery v0.21.0-alpha.0.0.20210609115025-669b54a1e5ed
	k8s.io/client-go v0.20.6
	k8s.io/klog/v2 v2.9.0
)

replace sigs.k8s.io/cluster-api-provider-aws => github.com/openshift/cluster-api-provider-aws v0.2.1-0.20201125052318-b85a18cbf338

replace sigs.k8s.io/cluster-api-provider-azure => github.com/openshift/cluster-api-provider-azure v0.0.0-20210209143830-3442c7a36c1e
