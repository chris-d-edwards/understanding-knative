package predicate_test

import (
	coordintatorPredicate "coordinator-watcher/controllers/predicate"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("ControllerPredicate", func() {
	var deployment *appsv1.Deployment
	var deploymentOld *appsv1.Deployment
	var oldReplicas int32 = 0
	var newReplicas int32 = 1
	BeforeEach(func() {
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: "biz", Name: "baz"},
			Spec: appsv1.DeploymentSpec{
				Replicas: &newReplicas,
			},
		}

		deploymentOld = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: "biz", Name: "baz"},
			Spec: appsv1.DeploymentSpec{
				Replicas: &oldReplicas,
			},
		}
	})

	Describe("Call Funcs", func() {

		coordinatorFunct := coordintatorPredicate.CoordinatorPredicate
		It("call Create should be false ever", func(done Done) {
			instance := coordinatorFunct
			evt := event.CreateEvent{
				Object: deployment,
			}
			Expect(instance.Create(evt)).To(BeFalse())
			close(done)
		})
		It("Update should be true when replicate scale from 0 to 1", func(done Done) {

			envUpdate := event.UpdateEvent{
				ObjectOld: deploymentOld,
				ObjectNew: deployment,
			}
			Expect(coordinatorFunct.UpdateFunc(envUpdate)).ShouldNot(BeFalse())
			close(done)

		})

		It("Update should be called when replicas is downScale", func(done Done) {

			var newReplicas = int32(0)
			var oldReplicas = int32(1)

			deployment.Spec.Replicas = &newReplicas
			deploymentOld.Spec.Replicas = &oldReplicas

			envUpdate := event.UpdateEvent{
				ObjectOld: deploymentOld,
				ObjectNew: deployment,
			}
			Expect(coordinatorFunct.UpdateFunc(envUpdate)).Should(BeTrue())

			close(done)
		})
	})

})
