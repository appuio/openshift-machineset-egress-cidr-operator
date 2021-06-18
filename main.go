package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/appuio/openshift-machineset-egress-cidr-operator/pkg/controller"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

const (
	leaseLockName = "machineset-egress-cidr-operator.appuio.ch"
)

func main() {
	// Parse command line flags and initialize logger
	klog.InitFlags(flag.CommandLine)
	flag.Parse()
	klog.Info("Starting up...")

	// load config from ServiceAccount or $KUBECONFIG file
	config := newConfig()
	ctrl := controller.New(config)

	// set up infrastructure for stopping stuff
	// stopCh will be passed to the controller to signal termination
	stopCh := make(chan struct{})
	defer close(stopCh)

	// ctx will be passed to LeaderLock to signal termination
	ctx, cancel := context.WithCancel(context.Background())
	// defers are called in order, it's important that stopCh is closed before
	// cancel is called since cancel will release the leaderlock.
	defer cancel()

	// Listen for OS signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		klog.Info("Exiting...")
		os.Exit(0)
		// defer calls will be fired
	}()

	// Configure leader lock
	// leaseIdentity must be unique for each started process
	leaseIdentity := uuid.New().String()
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseLockName,
			Namespace: getNamespace(),
		},
		Client: clientset.NewForConfigOrDie(config).CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: leaseIdentity,
		},
	}

	//
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Name: "openshift-meco-leader",
		Lock: lock,
		// Important: All code MUST be stopped BEFORE cancel is called!
		ReleaseOnCancel: true,

		// Use values from core clients
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,

		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				klog.Infof("Leader<%s>: got lease", leaseIdentity)
				ctrl.Run(stopCh)
			},
			OnStoppedLeading: func() {
				klog.Infof("Leader<%s>: lost lease", leaseIdentity)
				close(stopCh)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				if identity == leaseIdentity {
					return
				}
				klog.Infof("Leader<%s>: new leader elected: %s",
					leaseIdentity, identity)
			},
		},
	})

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

func getNamespace() string {
	if ns := getNamespaceFromFile(); ns != "" {
		return ns
	}

	if ns := getNamespaceFromKubeconfig(); ns != "" {
		return ns
	}

	if ns := os.Getenv("NAMESPACE"); ns != "" {
		return ns
	}

	klog.Error("Could not determine namespace, set $NAMESPACE")
	os.Exit(1)
	return ""
}

func getNamespaceFromFile() string {
	content, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ""
	}

	return string(content)
}

func getNamespaceFromKubeconfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		return ""
	}

	cfg, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return ""
	}

	return cfg.Contexts[cfg.CurrentContext].Namespace
}
