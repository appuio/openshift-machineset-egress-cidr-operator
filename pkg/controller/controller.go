package controller

import (
	"context"

	v1 "github.com/openshift/client-go/network/clientset/versioned/typed/network/v1"
	network "github.com/openshift/client-go/network/informers/externalversions"
	networkInformers "github.com/openshift/client-go/network/informers/externalversions/network/v1"
	networkListers "github.com/openshift/client-go/network/listers/network/v1"
	machine "github.com/openshift/machine-api-operator/pkg/generated/informers/externalversions"
	machineInformers "github.com/openshift/machine-api-operator/pkg/generated/informers/externalversions/machine/v1beta1"
	machineListers "github.com/openshift/machine-api-operator/pkg/generated/listers/machine/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	AnnotationEgressCIDRS = "appuio.ch/egress-cidrs"
	LeaseLockName         = "machineset-egress-cidr-operator.appuio.ch"
	MachineNamespace      = "openshift-machine-api"
	MachinesetLabel       = "machine.openshift.io/cluster-api-machineset"
	RoleLabel             = "machine.openshift.io/cluster-api-machine-role"
)

type Controller struct {
	cidrs *CIDRMap

	machineInformerFactory machine.SharedInformerFactory
	networkInformerFactory network.SharedInformerFactory

	machineSetInformer machineInformers.MachineSetInformer
	machineInformer    machineInformers.MachineInformer
	hostSubNetInformer networkInformers.HostSubnetInformer

	machines         machineListers.MachineNamespaceLister
	machineSets      machineListers.MachineSetNamespaceLister
	hostSubnets      networkListers.HostSubnetLister
	hostSubnetClient v1.HostSubnetInterface

	config *rest.Config
}

func New(config *rest.Config) *Controller {
	c := &Controller{
		cidrs:  NewCIDRMap(),
		config: config,
	}

	c.createMachineInformer()
	c.createNetworkInformer()

	return c
}

func (c *Controller) Run(ctx context.Context) {
	// Doing the Machine(Set) sync first to ensure our CIDR cache is warmed up
	c.machineInformerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(),
		c.machineInformer.Informer().HasSynced,
		c.machineSetInformer.Informer().HasSynced,
	) {
		klog.Fatal("Failed to do initial Machine sync")
	}

	c.networkInformerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), c.hostSubNetInformer.Informer().HasSynced) {
		klog.Fatal("Failed to do initial Network sync")
	}
}
