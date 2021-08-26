package controllers

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestReconcile(t *testing.T) {

	setupManager(t)

}

func setupManager(t *testing.T) {

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	_ = prometheus.NewRegistry()
	// manager.New(cfg, manager.Options{
	// 	MetricsBindAddress: "0",
	// 	NewCache:           dynamiccache.New,
	// 	MapperProvider: func(c *rest.Config) (meta.RESTMapper, error) {
	// 		return apiutil.NewDynamicRESTMapper(c)
	// 	},
	// })

}
