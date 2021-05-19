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

var splitterCondSet = apis.NewLivingConditionSet(SplitterConditionReady, SplitterConditionSinkReady, SplitterConditionServiceReady)

const (
	SplitterConditionReady = apis.ConditionReady

	SplitterConditionSinkReady    apis.ConditionType = "SinkReady"
	SplitterConditionServiceReady apis.ConditionType = "SplitterServiceReady"
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (*Splitter) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Splitter")
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (s *Splitter) GetConditionSet() apis.ConditionSet {
	return splitterCondSet
}

// InitializeConditions sets the initial values to the conditions.
func (ss *SplitterStatus) InitializeConditions() {
	splitterCondSet.Manage(ss).InitializeConditions()
}

// MarkServiceUnavailable updates Splitter status with Splitter Service Not Ready condition
func (ss *SplitterStatus) MarkServiceUnavailable(name string) {
	splitterCondSet.Manage(ss).MarkFalse(
		SplitterConditionServiceReady,
		"SplitterServiceUnavailable",
		"Splitter Service %q is not ready.", name)
}

// MarkServiceAvailable updates Splitter status with Splitter Service Is Ready condition
func (ss *SplitterStatus) MarkServiceAvailable() {
	splitterCondSet.Manage(ss).MarkTrue(SplitterConditionServiceReady)
}

// MarkSinkUnavailable updates Splitter status with Sink Not Ready condition
func (ss *SplitterStatus) MarkSinkUnavailable() {
	condSet.Manage(ss).MarkFalse(
		ConditionSinkReady,
		"SinkUnavailable",
		"Sink is unavailable")
}

// MarkSinkAvailable updates Splitter status with Sink Is Ready condition
func (ss *SplitterStatus) MarkSinkAvailable() {
	condSet.Manage(ss).MarkTrue(ConditionSinkReady)
}
