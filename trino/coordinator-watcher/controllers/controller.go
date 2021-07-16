package controllers

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ┌────────────────────────┬───────────────────────────────┬───────────────────────────┐
// │                        │                               │                           │
// │    Coodinator          │       Workers                 │     Actions               │
// │    (replicas)          │      (replicas)               │                           │
// │                        │                               │                           │
// ├────────────────────────┼───────────────────────────────┼───────────────────────────┤
// │                        │                               │                           │
// │       0                │        gt 0                   │   worker down             │
// ├────────────────────────┼───────────────────────────────┼───────────────────────────┤
// │       0                │          0                    │     NOPE                  │
// ├────────────────────────┼───────────────────────────────┼───────────────────────────┤
// │       1                │          0                    │     if exists scale       │
// ├────────────────────────┼───────────────────────────────┼───────────────────────────┤
// │       1                │          0                    │    not exist create       │
// ├────────────────────────┼───────────────────────────────┼───────────────────────────┤
// │       1                │       0 < x < N               │         NOPE              │
// ├────────────────────────┼───────────────────────────────┼───────────────────────────┤
// │       1                │          n                    │       NOPE                │
// └────────────────────────┴───────────────────────────────┴───────────────────────────┘

type ReconcileDeployments struct {

	// client can be used to retrieve objects from the APIServer.
	Client client.Client
	Log    logr.Logger
	Sink   string
}

var _ reconcile.Reconciler = &ReconcileDeployments{}

func (r *ReconcileDeployments) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	r.Log.Info("start to reconcile")

	_, err := r.getWorkerDeployment(request.Namespace)

	if errors.IsNotFound(err) {
		baseLabel := labels.Set{
			"app":        "trinodb-worker",
			"controller": "trinodb-name",
		}
		workerDeployment, err := r.createDeploymentForWorker(baseLabel, int32(1), request.Namespace)
		if err != nil {
			r.Log.Error(err, "Problem in Deployment definition")
			return reconcile.Result{}, err
		}
		if err = r.Client.Create(context.TODO(), workerDeployment); err != nil {
			r.Log.Error(err, "Problem in Deployment Creation")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
		//Create Deployment
	} else {
		if err != nil {
			r.Log.Error(err, "Problem while retrieving Deployment")
			return reconcile.Result{}, err
		}
		r.Log.Info("tobe done")

		// found worker. Check coordinator. if coodinator deplyments is 0 and worker gt 1 than i need to downscale othe
	}
	if err != nil {
		r.Log.Error(err, "search TrinoDB worker problems")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDeployments) getWorkerDeployment(namespace string) (*v1.Deployment, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{
		"app":        "trinodb-worker",
		"controller": "trinodb-name", // come from controller configuration
	}))
	existingDeployment := &v1.DeploymentList{}

	err := r.Client.List(context.TODO(), existingDeployment,
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: selector,
		})

	if errors.IsNotFound(err) {
		return nil, errors.NewNotFound(v1.Resource("deploymentes"), "")
	}
	if err != nil {
		r.Log.Error(err, "failed to list existing TrinoDB deployments")
		return nil, err
	}
	if len(existingDeployment.Items) == 0 {
		return nil, errors.NewNotFound(v1.Resource("deploymentes"), "")
	}
	return &existingDeployment.Items[0], nil
}

func (r *ReconcileDeployments) getCoordinatorDeplyments(namespace string) (*v1.Deployment, error) {

	existingDeployment := &v1.DeploymentList{}

	return &existingDeployment.Items[0], nil
}

func (r *ReconcileDeployments) createDeploymentForWorker(lbls map[string]string, workerCount int32, namespace string) (*v1.Deployment, error) {
	podSpec, err := createTrinoPodSpec()
	if err != nil {
		return nil, err
	}
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			//GenerateName:    getWorkerdeployments(presto.Status.Uuid), // todo fbalicchia
			GenerateName:    "trinodb-worker",
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{}, //todo fbalicchia
			Labels:          lbls,
		},
		Spec: v1.DeploymentSpec{
			Replicas: func() *int32 { i := workerCount; return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: lbls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					//OwnerReferences: []metav1.OwnerReference{*getOwnerReference(trinoDB)}, //todo fbalicchia
					Namespace: namespace,
					Labels:    lbls,
				},
				Spec: *podSpec,
			},
		},
	}, nil
}

func createTrinoPodSpec() (*corev1.PodSpec, error) {

	return &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "trinodb-worker",
				Image:           "kind.local/trino-356:356-cuebiq-20210617.111101",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Ports:           []corev1.ContainerPort{{ContainerPort: 8080}},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "config-volume", MountPath: "/etc/trino/"},
					{Name: "catalog-volume", MountPath: "/etc/trino/catalog"},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "config-volume",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "trinodb-worker",
						},
					},
				},
			},
			{
				Name: "catalog-volume",
				VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
					SecretName: "trino-connectors",
				}},
			},
		},
	}, nil
}
