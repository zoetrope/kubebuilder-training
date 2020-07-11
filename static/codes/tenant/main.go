package main

import (
	"flag"
	"os"
	"time"

	"github.com/go-logr/logr"
	multitenancyv1 "github.com/zoetrope/kubebuilder-training/static/codes/api/v1"
	"github.com/zoetrope/kubebuilder-training/static/codes/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = multitenancyv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "probe-addr", ":9090", "The address the liveness probe and readiness probe endpoints bind to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "27475f02.example.com",
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.TenantReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Tenant"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenant")
		os.Exit(1)
	}
	if err = (&multitenancyv1.Tenant{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Tenant")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	err = mgr.Add(&runner{ctrl.Log.WithName("runner")})
	if err != nil {
		setupLog.Error(err, "unable to add runner")
		os.Exit(1)
	}

	err = mgr.AddHealthzCheck("ping", healthz.Ping)
	if err != nil {
		setupLog.Error(err, "unable to add healthz check")
		os.Exit(1)
	}
	err = mgr.AddReadyzCheck("ping", healthz.Ping)
	if err != nil {
		setupLog.Error(err, "unable to add readyz check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

type runner struct {
	log logr.Logger
}

func (r runner) Start(ch <-chan struct{}) error {
	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ch:
			return nil
		case <-timer.C:
			r.log.Info("run something")
		}
	}
}

func (r runner) NeedLeaderElection() bool {
	return true
}
