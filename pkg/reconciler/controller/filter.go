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

package controller

import (
	"context"
	"fmt"

	corev1listers "k8s.io/client-go/listers/core/v1"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"
	"knative.dev/pkg/tracker"

	filterv1alpha1 "github.com/triggermesh/filter/pkg/apis/filter/v1alpha1"
	filterreconciler "github.com/triggermesh/filter/pkg/client/generated/injection/reconciler/filter/v1alpha1/filter"
)

const (
	eventType = "io.triggermesh.routing.filter"
)

// Reconciler implements addressableservicereconciler.Interface for
// AddressableService resources.
type Reconciler struct {
	// Tracker builds an index of what resources are watching other resources
	// so that we can immediately react to changes tracked resources.
	Tracker tracker.Interface

	// Listers index properties about resources
	ServiceLister corev1listers.ServiceLister

	sinkResolver *resolver.URIResolver

	Filter FilterService
}

// Check that our Reconciler implements Interface
var _ filterreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *filterv1alpha1.Filter) reconciler.Event {
	logger := logging.FromContext(ctx)

	if err := r.Tracker.TrackReference(tracker.Reference{
		APIVersion: "v1",
		Kind:       "Service",
		Name:       r.Filter.Name,
		Namespace:  r.Filter.Namespace,
	}, o); err != nil {
		logger.Errorf("Error tracking service %v: %v", r.Filter, err)
		return err
	}

	if _, err := r.ServiceLister.Services(r.Filter.Namespace).Get(r.Filter.Name); err != nil {
		logger.Errorf("Error reconciling service %v: %v", r.Filter, err)
		o.Status.MarkServiceUnavailable(fmt.Sprintf("%s/%s", r.Filter.Namespace, r.Filter.Name))
		return err
	}

	url, err := apis.ParseURL(fmt.Sprintf("http://%s/filters/%s/%s",
		network.GetServiceHostname(r.Filter.Name, r.Filter.Namespace), o.Namespace, o.Name))
	if err != nil {
		logger.Errorf("Error parsing service URL %v: %v", r.Filter, err)
		o.Status.MarkServiceUnavailable(fmt.Sprintf("%s/%s", r.Filter.Namespace, r.Filter.Name))
		return err
	}

	if o.Spec.Sink.Ref != nil && o.Spec.Sink.Ref.Namespace == "" {
		o.Spec.Sink.Ref.Namespace = o.Namespace
	}

	sink, err := r.sinkResolver.URIFromDestinationV1(ctx, *o.Spec.Sink, o)
	if err != nil {
		logger.Errorf("Error resolving sink URI: %v", err)
		o.Status.MarkSinkUnavailable()
		return err
	}

	o.Status.MarkSinkAvailable()
	o.Status.SinkURI = sink
	o.Status.CloudEventAttributes = []duckv1.CloudEventAttributes{
		{
			Type:   eventType,
			Source: o.SelfLink,
		},
	}

	o.Status.MarkServiceAvailable()
	o.Status.Address = &duckv1.Addressable{
		URL: url,
	}

	return nil
}
