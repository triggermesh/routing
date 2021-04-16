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

package splitter

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

	routingv1alpha1 "github.com/triggermesh/routing/pkg/apis/routing/v1alpha1"
	splitterreconciler "github.com/triggermesh/routing/pkg/client/generated/injection/reconciler/routing/v1alpha1/splitter"
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

	Splitter SplitterEnv
}

// Check that our Reconciler implements Interface
var _ splitterreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *routingv1alpha1.Splitter) reconciler.Event {
	logger := logging.FromContext(ctx)

	if err := r.Tracker.TrackReference(tracker.Reference{
		APIVersion: "v1",
		Kind:       "Service",
		Name:       r.Splitter.Name,
		Namespace:  r.Splitter.Namespace,
	}, o); err != nil {
		logger.Errorf("Error tracking service %v: %v", r.Splitter, err)
		return err
	}

	if _, err := r.ServiceLister.Services(r.Splitter.Namespace).Get(r.Splitter.Name); err != nil {
		logger.Errorf("Error reconciling service %v: %v", r.Splitter, err)
		o.Status.MarkServiceUnavailable(fmt.Sprintf("%s/%s", r.Splitter.Namespace, r.Splitter.Name))
		return err
	}

	url, err := apis.ParseURL(fmt.Sprintf("http://%s/splitters/%s/%s",
		network.GetServiceHostname(r.Splitter.Name, r.Splitter.Namespace), o.Namespace, o.Name))
	if err != nil {
		logger.Errorf("Error parsing service URL %v: %v", r.Splitter, err)
		o.Status.MarkServiceUnavailable(fmt.Sprintf("%s/%s", r.Splitter.Namespace, r.Splitter.Name))
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
			Type:   o.Spec.CEContext.Type,
			Source: o.Spec.CEContext.Source,
		},
	}

	o.Status.MarkServiceAvailable()
	o.Status.Address = &duckv1.Addressable{
		URL: url,
	}

	return nil
}
