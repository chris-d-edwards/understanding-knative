package knative

import (
	"context"
	"fmt"

	lazyapiv1 "lazy-service-controller/api/v1"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"knative.dev/pkg/kmp"
	"knative.dev/serving/pkg/apis/autoscaling"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type KsvcReconciler struct {
	client        client.Client
	Service       *knservingv1.Service
	targetService *lazyapiv1.Lazyservice
	componentExt  *lazyapiv1.ComponentExtensionSpec
	scheme        *runtime.Scheme
}

func NewKsvcReconciler(client client.Client, componentMeta metav1.ObjectMeta,
	componentExt *lazyapiv1.ComponentExtensionSpec, scheme *runtime.Scheme,
	targetService *lazyapiv1.Lazyservice,
) *KsvcReconciler {
	return &KsvcReconciler{
		client:  client,
		Service: createKnativeService(targetService, componentMeta, componentExt, scheme),
	}
}

func createKnativeService(targetService *lazyapiv1.Lazyservice, componentMeta metav1.ObjectMeta,
	componentExtension *lazyapiv1.ComponentExtensionSpec, scheme *runtime.Scheme) *knservingv1.Service {

	annotations := componentMeta.Annotations
	annotations[autoscaling.MinScaleAnnotationKey] = fmt.Sprint(0)

	//componentExtension.MaxReplicas 2
	if componentExtension.MaxReplicas != 0 {
		annotations[autoscaling.MaxScaleAnnotationKey] = fmt.Sprint(componentExtension.MaxReplicas)
	}

	service := &knservingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      componentMeta.Name,
			Namespace: componentMeta.Namespace,
			Labels:    componentMeta.Labels,
			//OwnerReferences: addOwnerReference(targetService), //fbalicchia necessary to use controllerUtils controllerutil.SetOwnerReference(targetService, service, scheme.Scheme);
		},
		Spec: knservingv1.ServiceSpec{
			ConfigurationSpec: knservingv1.ConfigurationSpec{
				Template: knservingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      componentMeta.Labels,
						Annotations: annotations,
					},
					Spec: knservingv1.RevisionSpec{
						TimeoutSeconds: componentExtension.TimeoutSeconds,
						PodSpec: v1.PodSpec{
							InitContainers: []v1.Container{},
							Containers: []v1.Container{
								{
									Name:  targetService.ObjectMeta.Name,
									Image: targetService.Spec.Image,
								},
							}},
					},
				},
			},
		},
	}
	//Call setDefaults on desired knative service here to avoid diffs generated because knative defaulter webhook is
	//called when creating or updating the knative service
	if err := controllerutil.SetOwnerReference(targetService, service, scheme); err != nil {
		return nil
	}
	service.SetDefaults(context.TODO())
	return service
}

//AddOwnerReference add reference of CRD in hierarchy
func addOwnerReference(targetService *lazyapiv1.Lazyservice) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(targetService, schema.GroupVersionKind{
			Group:   lazyapiv1.GroupVersion.Group,
			Version: lazyapiv1.GroupVersion.Version,
			Kind:    "Lazyservice",
		}),
	}
}

func (r *KsvcReconciler) Reconcile() error {
	desired := r.Service
	existing := &knservingv1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if err != nil {
		if apierr.IsNotFound(err) {
			log.Info("Creating knative service", "namespace", desired.Namespace, "name", desired.Name)
			if err := r.client.Create(context.TODO(), desired); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if semanticEquals(desired, existing) {
		return nil
	}

	// Reconcile differences and update
	diff, err := kmp.SafeDiff(desired.Spec.ConfigurationSpec, existing.Spec.ConfigurationSpec)
	if err != nil {
		return errors.Wrapf(err, "failed to diff knative service configuration spec")
	}
	log.Info("knative service configuration diff (-desired, +observed):", "diff", diff)
	existing.Spec.ConfigurationSpec = desired.Spec.ConfigurationSpec
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels

	diff, err = kmp.SafeDiff(desired.Spec.RouteSpec, existing.Spec.RouteSpec)
	if err != nil {
		return errors.Wrapf(err, "fails to diff knative service route spec")
	}
	if diff != "" {
		log.Info("knative service routing spec diff (-desired, +observed):", "diff", diff)
	}
	existing.Spec.Traffic = desired.Spec.Traffic

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		log.Info("Updating knative service", "namespace", desired.Namespace, "name", desired.Name)
		return r.client.Update(context.TODO(), existing)
	})
	if err != nil {
		return errors.Wrapf(err, "fails to update knative service")
	}

	return nil
}

func semanticEquals(desiredService, service *knservingv1.Service) bool {
	return equality.Semantic.DeepEqual(desiredService.Spec.ConfigurationSpec, service.Spec.ConfigurationSpec) &&
		equality.Semantic.DeepEqual(desiredService.ObjectMeta.Labels, service.ObjectMeta.Labels) &&
		equality.Semantic.DeepEqual(desiredService.Spec.RouteSpec, service.Spec.RouteSpec)
}
