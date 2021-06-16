package controller_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/appuio/openshift-machineset-egress-cidr-operator/pkg/controller"
	"github.com/go-logr/logr"
	"github.com/matryer/is"
	v1 "github.com/openshift/api/network/v1"
	"github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// TEST-NET-1 192.0.2.0/24
// TEST-NET-2 198.51.100.0/24
// TEST-NET-3 203.0.113.0/24

func init() {
	klog.LogToStderr(false)
	klog.SetOutput(io.Discard)
	klog.SetLogger(logr.Discard())
}

func TestReconcileNoEntry(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	cm := controller.NewCIDRMap()
	getMachine, getMachineCalled := mockGetMachine(t, "some", hs.Name)

	is.Equal(controller.ReconcileSubnet(hs, cm, getMachine, nil), "no cidr entry")

	is.Equal(*getMachineCalled, 1) // getMachine called exactly once
}

func TestReconcileUpToDate(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	hs.EgressCIDRs = []v1.HostSubnetEgressCIDR{"192.0.2.0/24"}
	cm := controller.NewCIDRMap()
	getMachine, getMachineCalled := mockGetMachine(t, "some", hs.Name)
	cm.Set("some", "192.0.2.0/24")

	is.Equal(controller.ReconcileSubnet(hs, cm, getMachine, nil), "up to date")

	is.Equal(*getMachineCalled, 1) // getMachine called exactly once
}

func TestReconcileSetNew(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	cm := controller.NewCIDRMap()

	getMachine, getMachineCalled := mockGetMachine(t, "some", hs.Name)
	updateHostSubnet, updateHostSubnetCalled := mockUpdateHostSubnet(t, []v1.HostSubnetEgressCIDR{"192.0.2.0/24"})

	cm.Set("some", "192.0.2.0/24")
	is.Equal(controller.ReconcileSubnet(hs, cm, getMachine, updateHostSubnet), "updated")

	is.Equal(*getMachineCalled, 1)       // getMachine called exactly once
	is.Equal(*updateHostSubnetCalled, 1) // updateHostSubnet called exactly once
}

func TestReconcileUpdate(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	hs.EgressCIDRs = []v1.HostSubnetEgressCIDR{"192.0.2.0/24"}
	cm := controller.NewCIDRMap()

	getMachine, getMachineCalled := mockGetMachine(t, "some", hs.Name)
	updateHostSubnet, updateHostSubnetCalled := mockUpdateHostSubnet(t,
		[]v1.HostSubnetEgressCIDR{"198.51.100.0/24", "203.0.113.0/24"})

	cm.Set("some", "203.0.113.0/24,198.51.100.0/24")
	is.Equal(controller.ReconcileSubnet(hs, cm, getMachine, updateHostSubnet), "updated")

	is.Equal(*getMachineCalled, 1)       // getMachine called exactly once
	is.Equal(*updateHostSubnetCalled, 1) // updateHostSubnet called exactly once
}

func TestReconcileErrGetMachine(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	counter := 0

	getMachine := func(name string) (*v1beta1.Machine, error) {
		counter++
		return nil, errors.New("oh noes")
	}

	is.Equal(controller.ReconcileSubnet(hs, nil, getMachine, nil), "error getMachine: oh noes")
	is.Equal(counter, 1)
}

func TestReconcileNoMachineset(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	getMachine, getMachineCalled := mockGetMachine(t, "", hs.Name)

	is.Equal(controller.ReconcileSubnet(hs, nil, getMachine, nil), "error: no machineset label")
	is.Equal(*getMachineCalled, 1)
}

func TestReconcileNoCIDREntry(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	cm := controller.NewCIDRMap()
	getMachine, getMachineCalled := mockGetMachine(t, "aaa", hs.Name)

	is.Equal(controller.ReconcileSubnet(hs, cm, getMachine, nil), "no cidr entry")
	is.Equal(*getMachineCalled, 1)
}

func TestReconcileErrUpdateHostsubnet(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("node123")
	cm := controller.NewCIDRMap()
	cm.Set("aaa", "203.0.113.0/24")
	getMachine, getMachineCalled := mockGetMachine(t, "aaa", hs.Name)
	updateHostSubnetCalled := 0
	updateHostSubnet := func(ctx context.Context, hostSubnet *v1.HostSubnet, opts metav1.UpdateOptions) (*v1.HostSubnet, error) {
		updateHostSubnetCalled++
		return nil, errors.New("some oopsie")
	}

	is.Equal(
		controller.ReconcileSubnet(hs, cm, getMachine, updateHostSubnet),
		"error update hostsubnet: some oopsie",
	)

	is.Equal(*getMachineCalled, 1)
	is.Equal(updateHostSubnetCalled, 1)
}

func TestReconcileIgnoreMaster(t *testing.T) {
	is := is.New(t)
	hs := mockHostSubnet("master-asdf")
	counter := 0

	getMachine := func(name string) (*v1beta1.Machine, error) {
		m := new(v1beta1.Machine)
		m.SetLabels(map[string]string{
			controller.RoleLabel: "master",
		})
		return m, nil
	}

	is.Equal(controller.ReconcileSubnet(hs, nil, getMachine, nil), "ignore master")
	is.Equal(counter, 0) // getMachine called exactly once
}

func mockHostSubnet(name string) *v1.HostSubnet {
	hs := v1.HostSubnet{}
	hs.SetName(name)
	return &hs
}

func mockGetMachine(t *testing.T, machineset, expectedName string) (controller.MachineGetter, *int) {
	is := is.New(t)
	counter := new(int)

	m := new(v1beta1.Machine)
	if machineset != "" {
		m.SetLabels(map[string]string{
			controller.MachinesetLabel: machineset,
		})
	}

	fn := func(name string) (*v1beta1.Machine, error) {
		(*counter)++
		is.Equal(expectedName, name)
		return m, nil
	}

	return fn, counter
}

func mockUpdateHostSubnet(t *testing.T, expectedHostsubnet []v1.HostSubnetEgressCIDR) (controller.HostSubnetUpdater, *int) {
	is := is.New(t)
	counter := new(int)

	fn := func(ctx context.Context, hostSubnet *v1.HostSubnet, opts metav1.UpdateOptions) (*v1.HostSubnet, error) {
		(*counter)++
		is.Equal(hostSubnet.EgressCIDRs, expectedHostsubnet)
		return hostSubnet, nil
	}

	return fn, counter
}
