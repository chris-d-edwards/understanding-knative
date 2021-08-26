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

package controllers

import (
	"context"

	"lazy-service-controller/controllers/components"

	deployv1 "lazy-service-controller/api/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LazyserviceReconciler reconciles a Lazyservice object
type LazyserviceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=deploy.balyfix.ai,resources=lazyservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=deploy.balyfix.ai,resources=lazyservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=deploy.balyfix.ai,resources=lazyservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=serving.knative.dev,resources=services/finalizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=serving.knative.dev,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices/finalizers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Lazyservice object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *LazyserviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("lazyservice", req.NamespacedName)

	var lazyService deployv1.Lazyservice

	if err := r.Get(ctx, req.NamespacedName, &lazyService); err != nil {
		//log.Error(err, "unable to fetch lazyService")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	app := components.NewTargetApp(r.Client, r.Scheme)
	if err := app.Reconcile(&lazyService); err != nil {
		log.Error(err, "unable to reconcile service")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LazyserviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deployv1.Lazyservice{}).
		Complete(r)
}
