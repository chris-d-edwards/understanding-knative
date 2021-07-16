package predicate

import (
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var CoordinatorPredicate = predicate.Funcs{
	CreateFunc: func(event.CreateEvent) bool {
		return false
	},
	DeleteFunc: func(event.DeleteEvent) bool {
		return false
	},
	UpdateFunc: func(event event.UpdateEvent) bool {
		newObj := event.ObjectNew.(*v1.Deployment)
		oldObj := event.ObjectOld.(*v1.Deployment)

		newReplicas := newObj.Spec.Replicas
		oldReplicas := oldObj.Spec.Replicas
		baseValue := int32(0)
		aspectedValue := int32(1)

		isUPscale := *oldReplicas == baseValue && *newReplicas >= aspectedValue
		isDownscale := *oldReplicas == aspectedValue && *newReplicas <= aspectedValue

		return isUPscale || isDownscale
	},
	GenericFunc: func(event.GenericEvent) bool {

		return false
	},
}
