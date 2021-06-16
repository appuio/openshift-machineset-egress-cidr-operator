package controller_test

import (
	"testing"

	"github.com/appuio/openshift-machineset-egress-cidr-operator/pkg/controller"
	"github.com/matryer/is"
	v1 "github.com/openshift/api/network/v1"
)

func TestCIDRMap(t *testing.T) {
	is := is.New(t)
	cm := controller.NewCIDRMap()
	is.True(cm != nil)

	for _, c := range []struct {
		Name, Input string
		Expected    []v1.HostSubnetEgressCIDR
	}{
		{"simple", "foo", []v1.HostSubnetEgressCIDR{"foo"}},
		{"no-space", "foo,bar", []v1.HostSubnetEgressCIDR{"bar", "foo"}},
		{"one-space", "ggg, aaa", []v1.HostSubnetEgressCIDR{"aaa", "ggg"}},
		{"multiple-spaces", "one,   two", []v1.HostSubnetEgressCIDR{"one", "two"}},
		{"tabs", "tab,	ulator", []v1.HostSubnetEgressCIDR{"tab", "ulator"}},
	} {
		t.Run(c.Name, func(t *testing.T) {
			is := is.New(t)

			cm.Set(c.Name, c.Input)

			is.Equal(c.Expected, cm.Get(c.Name))
			is.True(cm.Exists(c.Name))
			is.True(cm.EqualCIRDs(c.Name, c.Expected))
			is.True(cm.Equals(c.Name, c.Input))

			cm.Delete(c.Name)
			is.True(!cm.Exists(c.Name))
			is.True(cm.Equals(c.Name, ""))
		})
	}
}

func TestCIDRMapEmptyString(t *testing.T) {
	is := is.New(t)
	cm := controller.NewCIDRMap()

	cm.Set("foo", "")
	is.True(!cm.Exists("foo"))
}
