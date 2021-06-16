package controller

import (
	"time"

	v1 "github.com/openshift/api/network/v1"
	"github.com/openshift/client-go/network/clientset/versioned"
	"github.com/openshift/client-go/network/informers/externalversions"
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
	ReconcileSubnet(hs, c.cidrs, c.machines.Get, c.hostSubnetClient.Update)
}

func (c *Controller) UpdateHostSubnet(_, hs *v1.HostSubnet) {
	ReconcileSubnet(hs, c.cidrs, c.machines.Get, c.hostSubnetClient.Update)
}

func (c *Controller) DeleteHostSubnet(hs *v1.HostSubnet) {}
