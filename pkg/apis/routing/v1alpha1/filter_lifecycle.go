/*
Copyright 2021 Triggermesh Inc.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var condSet = apis.NewLivingConditionSet(ConditionReady, ConditionSinkReady, ConditionFilterReady)

const (
	ConditionReady = apis.ConditionReady

	ConditionSinkReady apis.ConditionType = "SinkReady"

	ConditionFilterReady apis.ConditionType = "FilterServiceReady"
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (*Filter) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Filter")
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (f *Filter) GetConditionSet() apis.ConditionSet {
	return condSet
}

// InitializeConditions sets the initial values to the conditions.
func (fs *FilterStatus) InitializeConditions() {
	condSet.Manage(fs).InitializeConditions()
}

// MarkServiceUnavailable updates Filter status with Filter Service Not Ready condition
func (fs *FilterStatus) MarkServiceUnavailable(name string) {
	condSet.Manage(fs).MarkFalse(
		ConditionFilterReady,
		"FilterServiceUnavailable",
		"Filter Service %q is not ready.", name)
}

// MarkServiceAvailable updates Filter status with Filter Service Is Ready condition
func (fs *FilterStatus) MarkServiceAvailable() {
	condSet.Manage(fs).MarkTrue(ConditionFilterReady)
}

// MarkSinkUnavailable updates Filter status with Sink Not Ready condition
func (fs *FilterStatus) MarkSinkUnavailable() {
	condSet.Manage(fs).MarkFalse(
		ConditionSinkReady,
		"SinkUnavailable",
		"Sink is unavailable")
}

// MarkSinkAvailable updates Filter status with Sink Is Ready condition
func (fs *FilterStatus) MarkSinkAvailable() {
	condSet.Manage(fs).MarkTrue(ConditionSinkReady)
}
