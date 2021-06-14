package controller

import (
	"time"

	"github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	"github.com/openshift/machine-api-operator/pkg/generated/clientset/versioned"
	"github.com/openshift/machine-api-operator/pkg/generated/informers/externalversions"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (c *Controller) createMachineInformer() {
	clientset, err := versioned.NewForConfig(c.config)
	if err != nil {
		klog.Fatalln(err)
	}

	// what is this, Java?
	factory := externalversions.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Hour,
		externalversions.WithNamespace(MachineNamespace),
	)
	machineInformer := factory.Machine().V1beta1().Machines()
	machineSetInformer := factory.Machine().V1beta1().MachineSets()
	machineSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ms := obj.(*v1beta1.MachineSet)
			c.AddMachineSet(ms)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldMs := oldObj.(*v1beta1.MachineSet)
			newMs := newObj.(*v1beta1.MachineSet)
			c.UpdateMachineSet(oldMs, newMs)
		},
		DeleteFunc: func(obj interface{}) {
			ms := obj.(*v1beta1.MachineSet)
			c.DeleteMachineSet(ms)
		},
	})

	c.machineInformerFactory = factory
	c.machineSetInformer = machineSetInformer
	c.machineInformer = machineInformer
	c.machines = machineInformer.Lister().Machines(MachineNamespace)
	c.machineSets = machineSetInformer.Lister().MachineSets(MachineNamespace)
}

func (c *Controller) AddMachineSet(ms *v1beta1.MachineSet) {
	cidrs := ms.Annotations[AnnotationEgressCIDRS]

	if cidrs == "" {
		c.cidrs.Delete(ms.Name)
		return
	}

	c.cidrs.Insert(ms.Name, cidrs)
}

func (c *Controller) UpdateMachineSet(_, ms *v1beta1.MachineSet) {
	cidrs := ms.Annotations[AnnotationEgressCIDRS]

	if cidrs == "" {
		c.cidrs.Delete(ms.Name)
		return
	}

	if !c.cidrs.Equals(ms.Name, cidrs) {
		c.cidrs.Insert(ms.Name, cidrs)
		return
	}
}

func (c *Controller) DeleteMachineSet(ms *v1beta1.MachineSet) {
	c.cidrs.Delete(ms.Name)
}
