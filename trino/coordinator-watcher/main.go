/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	localcontroller "coordinator-watcher/controllers"
	coordintatorPredicate "coordinator-watcher/controllers/predicate"

	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/source"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var sinkAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&sinkAddr, "trino-worker-address", "http://trino-worker.default.127.0.0.1.nip.io", "The trinoWorker address that need to awake ")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "7305b01c.balyfix.ai",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	setupLog.Info("Setting up controller")
	c, err := controller.New("trino-worker-wakeup-controller", mgr, controller.Options{
		Reconciler: &localcontroller.ReconcileDeployments{Client: mgr.GetClient(), Log: ctrl.Log.WithName("trino-wakeup-controller"), Sink: sinkAddr},
	})
	if err != nil {
		setupLog.Error(err, "unable to set up individual controller")
		os.Exit(1)
	}

	// Watch Deploymentes and enqueue Deployments object key

	labelToWatch := map[string]string{
		"app.kubernetes.io/component": "coordinator",
	}

	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labelToWatch,
		},
	}}, &handler.EnqueueRequestForObject{}, coordintatorPredicate.CoordinatorPredicate); err != nil {
		setupLog.Error(err, "unable to watch Deployments")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
