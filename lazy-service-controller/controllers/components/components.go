package components

import v1 "lazy-service-controller/api/v1"

type Component interface {
	Reconcile(service *v1.Lazyservice) error
}
