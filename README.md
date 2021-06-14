# OpenShift MachineSet EgressCIDR operator

Automatically configure `egressCIDRs` for all nodes belonging to a MachineSet.

## Usage

Annotate your MachineSet with a comma-separated list of CIDRs you want to have on your nodes:

    oc annotate machineset/foo appuio.ch/egress-cidrs=192.0.2.0/27,192.0.2.128/27

This will OVERRIDE the `egressCIDRs` field for all HostSubnets belonging to the Machineset.

## Development

Apply all the manifests in `manifests/` to your test cluster:

    kubectl apply -k ./manifests

Fetch the kubeconfig file:

    oc -n appuio-machineset-egress-cidr-operator sa create-kubeconfig operator > operator.kubeconfig

You can now run the operator out-of-cluster:

    KUBECONFIG=$(pwd)/operator.kubeconfig go run .

### Update dependencies

Run `make deps` to fetch the latest dependencies. Upgrade the version Numbers on top of `Makefile` if you want to upgrade to a newer OpenShift relase. Make sure the K8s version matches the OCP version! Check the OCP release notes if you are unsure.

## License

BSD-3-Clause](LICENSE)
