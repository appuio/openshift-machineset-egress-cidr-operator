package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/appuio/openshift-machineset-egress-cidr-operator/pkg/controller"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(flag.CommandLine)
	flag.Parse()
	klog.Info("Starting up...")

	config := newConfig()
	ctrl := controller.New(config)

	stopCh := make(chan struct{})
	defer close(stopCh)

	ctrl.Run(stopCh)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	klog.Info("Exiting...")
}

func newConfig() *rest.Config {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		klog.Infof("KUBECONFIG is set, using %s", kubeconfig)
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Fatal(err)
		}
		return config
	}

	klog.Info("using in-cluster config")
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatal(err)
	}

	return config
}
