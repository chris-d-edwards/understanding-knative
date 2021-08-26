package components

import (
	v1 "lazy-service-controller/api/v1"
	"lazy-service-controller/controllers/reconcilers/knative"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//Compiler verify that TargetApp implementes Components
//https://golang.org/doc/faq#guarantee_satisfies_interface
var _ Component = &TargetApp{}

type TargetApp struct {
	client client.Client
	scheme *runtime.Scheme
	Log    logr.Logger
}

func NewTargetApp(client client.Client, scheme *runtime.Scheme) Component {
	return &TargetApp{
		client: client,
		scheme: scheme,
		Log:    ctrl.Log.WithName("TargetAppReconciler"),
	}
}

func (t *TargetApp) Reconcile(service *v1.Lazyservice) error {

	objectMeta := metav1.ObjectMeta{
		Name:      service.ObjectMeta.Name,
		Namespace: "default", // understand how obtain namespace
		Labels: map[string]string{
			"label1": "valueLabel1",
			"label2": "valueLabel1",
		},
		Annotations: map[string]string{},
	}

	r := knative.NewKsvcReconciler(t.client, objectMeta, &service.Spec.ComponentExtensionSpec, t.scheme, service)

	if err := r.Reconcile(); err != nil {
		return err
	}

	return nil
}
