# OpenShift MachineSet EgressCIDR operator

[![Docker Repository on Quay](https://quay.io/repository/appuio/openshift-machineset-egress-cidr-operator/status "Docker Repository on Quay")](https://quay.io/repository/appuio/openshift-machineset-egress-cidr-operator)

Automatically configure `egressCIDRs` for all nodes belonging to a MachineSet.

## Usage

Annotate your MachineSet with a comma-separated list of CIDRs you want to have on your nodes:

    oc annotate machineset/foo appuio.ch/egress-cidrs=192.0.2.0/27,192.0.2.128/27

This will OVERRIDE the `egressCIDRs` field for all HostSubnets belonging to the Machineset.

Removing the annotation will make the field unmanaged. To explicitly REMOVE any `egressCIDRs`, set the annotation to the value `"none"`.

    oc annotate machineset/foo appuio.ch/egress-cidrs=none

## Deployment

When running the operator in-cluster, it will autodiscover the service account. When running out of cluster, make sure to set the `KUBECONFIG` env var.

You can configure logging verbosity by using the `-v` flag, however note that this will be applied to the whole K8s client library. To only get relevant stuff, uset `-vmodule=reconcile=8`.

## Development

Apply all the manifests in `manifests/` to your test cluster:

    oc create namespace appuio-machineset-egress-cidr-operator
    kubectl apply -k ./manifests

Fetch the kubeconfig file:

    oc -n appuio-machineset-egress-cidr-operator sa create-kubeconfig operator > operator.kubeconfig

You can now run the operator out-of-cluster:

    KUBECONFIG=$(pwd)/operator.kubeconfig go run . -vmodule=reconcile=8

### Update dependencies

Run `make deps` to fetch the latest dependencies. Upgrade the version Numbers on top of `Makefile` if you want to upgrade to a newer OpenShift relase. Make sure the K8s version matches the OCP version! Check the OCP release notes if you are unsure.

## License

[BSD-3-Clause](LICENSE)
