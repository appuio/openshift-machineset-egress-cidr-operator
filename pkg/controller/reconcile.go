package controller

import (
	"context"

	v1 "github.com/openshift/api/network/v1"
	"github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type MachineGetter func(name string) (*v1beta1.Machine, error)
type HostSubnetUpdater func(ctx context.Context, hostSubnet *v1.HostSubnet, opts metav1.UpdateOptions) (*v1.HostSubnet, error)

func ReconcileSubnet(
	hs *v1.HostSubnet,
	cidrs *CIDRMap,
	getMachine MachineGetter,
	updateHostSubnet HostSubnetUpdater,
) string {
	klog.V(8).Infof("HostSubnet<%s>: Reconcile", hs.Name)
	machine, err := getMachine(hs.Name)
	if err != nil {
		klog.Errorf("HostSubnet<%s>: get machine: %s", hs.Name, err)
		return "error getMachine: " + err.Error()
	}

	if machine.Labels[RoleLabel] == "master" {
		klog.V(8).Infof("HostSubnet<%s>: role==master; ignore", hs.Name)
		return "ignore master"
	}

	machineset := machine.Labels[MachinesetLabel]
	if machineset == "" {
		klog.Errorf("HostSubnet<%s>: no '%s' label on machine", hs.Name, MachinesetLabel)
		return "error: no machineset label"
	}

	if !cidrs.Exists(machineset) {
		klog.V(8).Infof("HostSubnet<%s>: No or empty entry in CIDR cache, skipping", hs.Name)
		return "no cidr entry"
	}

	actual := hs.EgressCIDRs
	if cidrs.EqualCIRDs(machineset, actual) {
		klog.V(8).Infof("HostSubnet<%s>: Already matches desired value, skipping", hs.Name)
		return "up to date"
	}

	desired := cidrs.Get(machineset)
	klog.Infof("HostSubnet<%s>: Out of date, updating.", hs.Name)
	klog.Infof("HostSubnet<%s>: Old value: %v", hs.Name, actual)
	klog.Infof("HostSubnet<%s>: New value: %v", hs.Name, desired)
	hs.EgressCIDRs = desired
	_, err = updateHostSubnet(context.Background(), hs, metav1.UpdateOptions{
		FieldManager: "openshift-machineset-egress-cidr-operator",
	})
	if err != nil {
		klog.Error("HostSubnet<%s>: updating: %s", hs.Name, err)
		return "error update hostsubnet: " + err.Error()
	}

	return "updated"
}
