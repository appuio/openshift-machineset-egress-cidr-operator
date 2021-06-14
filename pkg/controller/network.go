package controller

import (
	"context"
	"time"

	v1 "github.com/openshift/api/network/v1"
	"github.com/openshift/client-go/network/clientset/versioned"
	"github.com/openshift/client-go/network/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (c *Controller) createNetworkInformer() {
	clientset, err := versioned.NewForConfig(c.config)
	if err != nil {
		klog.Fatal(err)
	}

	factory := externalversions.NewSharedInformerFactory(clientset, time.Hour)
	informer := factory.Network().V1().HostSubnets()
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ms := obj.(*v1.HostSubnet)
			c.AddHostSubnet(ms)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldMs := oldObj.(*v1.HostSubnet)
			newMs := newObj.(*v1.HostSubnet)
			c.UpdateHostSubnet(oldMs, newMs)
		},
		DeleteFunc: func(obj interface{}) {
			ms := obj.(*v1.HostSubnet)
			c.DeleteHostSubnet(ms)
		},
	})

	c.networkInformerFactory = factory
	c.hostSubNetInformer = informer
	c.hostSubnets = informer.Lister()
	c.hostSubnetClient = clientset.NetworkV1().HostSubnets()
}

func (c *Controller) AddHostSubnet(hs *v1.HostSubnet) {
	c.reconcileSubnet(hs)
}

func (c *Controller) UpdateHostSubnet(_, hs *v1.HostSubnet) {
	c.reconcileSubnet(hs)
}

func (c *Controller) DeleteHostSubnet(hs *v1.HostSubnet) {}

func (c *Controller) reconcileSubnet(hs *v1.HostSubnet) {
	klog.V(8).Infof("HostSubnet<%s>: Reconcile", hs.Name)
	machine, err := c.machines.Get(hs.Name)
	if err != nil {
		klog.Errorf("HostSubnet<%s>: get machine: %s", hs.Name, err)
		return
	}

	machineset, ok := machine.Labels[MachinesetLabel]
	if !ok {
		klog.Errorf("HostSubnet<%s>: no '%s' label on machine", hs.Name, MachinesetLabel)
		return
	}

	if !c.cidrs.Exists(machineset) {
		klog.V(8).Infof("HostSubnet<%s>: No or empty entry in CIDR cache, skipping", hs.Name)
		return
	}

	actual := hs.EgressCIDRs
	if c.cidrs.EqualCIRDs(machineset, actual) {
		klog.V(8).Infof("HostSubnet<%s>: Already matches desired value, skipping", hs.Name)
		return
	}

	desired := c.cidrs.Get(machineset)
	klog.Infof("HostSubnet<%s>: Out of date, updating.", hs.Name)
	klog.Infof("HostSubnet<%s>: Old value: %v", hs.Name, actual)
	klog.Infof("HostSubnet<%s>: New value: %v", hs.Name, desired)
	hs.EgressCIDRs = desired
	_, err = c.hostSubnetClient.Update(context.Background(), hs, metav1.UpdateOptions{
		FieldManager: "openshift-machineset-egress-cidr-operator",
	})
	if err != nil {
		klog.Error("HostSubnet<%s>: updating: %s", hs.Name, err)
	}
}
