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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// ComponentExtensionSpec defines the deployment configuration for a given InferenceService component
type ComponentExtensionSpec struct {
	// Minimum number of replicas, defaults to 1 but can be set to 0 to enable scale-to-zero.
	// +optional
	MinReplicas *int `json:"minReplicas,omitempty"`
	// Maximum number of replicas for autoscaling.
	// +optional
	MaxReplicas int `json:"maxReplicas,omitempty"`
	// ContainerConcurrency specifies how many requests can be processed concurrently, this sets the hard limit of the container
	// concurrency(https://knative.dev/docs/serving/autoscaling/concurrency).
	// +optional
	ContainerConcurrency *int64 `json:"containerConcurrency,omitempty"`
	// TimeoutSeconds specifies the number of seconds to wait before timing out a request to the component.
	// +optional
	TimeoutSeconds *int64 `json:"timeout,omitempty"`
	// CanaryTrafficPercent defines the traffic split percentage between the candidate revision and the last ready revision
	// +optional
	CanaryTrafficPercent *int64 `json:"canaryTrafficPercent,omitempty"`
}

// ComponentStatusSpec describes the state of the component
type ComponentStatusSpec struct {
	// Latest revision name that is in ready state
	// +optional
	LatestReadyRevision string `json:"latestReadyRevision,omitempty"`
	// Latest revision name that is created
	// +optional
	LatestCreatedRevision string `json:"latestCreatedRevision,omitempty"`
	// Previous revision name that is rolled out with 100 percent traffic
	// +optional
	PreviousRolledoutRevision string `json:"previousRolledoutRevision,omitempty"`
	// Latest revision name that is rolled out with 100 percent traffic
	// +optional
	//LatestRolledoutRevision string `json:"latestRolledoutRevision,omitempty"`
	// Traffic holds the configured traffic distribution for latest ready revision and previous rolled out revision.
	// +optional
	Traffic []knservingv1.TrafficTarget `json:"traffic,omitempty"`
	// URL holds the url that will distribute traffic over the provided traffic targets.
	// It generally has the form http[s]://{route-name}.{route-namespace}.{cluster-level-suffix}
	// +optional
	URL *apis.URL `json:"url,omitempty"`
	// Addressable endpoint for the InferenceService
	// +optional
	Address *duckv1.Addressable `json:"address,omitempty"`
}

// LazyserviceSpec defines the desired state of Lazyservice
type LazyserviceSpec struct {
	Foo                    string `json:"foo,omitempty"`
	ServiceName            string `json:"serviceName,omitemty"`
	Image                  string `json:"image,omitempty"`
	ComponentExtensionSpec `json:",inline"`
}

// LazyserviceStatus defines the observed state of Lazyservice
type LazyserviceStatus struct {
	ComponentStatus ComponentStatusSpec `json:"componentStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Lazyservice is the Schema for the lazyservices API
type Lazyservice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LazyserviceSpec   `json:"spec,omitempty"`
	Status LazyserviceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LazyserviceList contains a list of Lazyservice
type LazyserviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Lazyservice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Lazyservice{}, &LazyserviceList{})
}
