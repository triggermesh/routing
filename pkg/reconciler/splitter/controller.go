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

	corev1 "k8s.io/api/core/v1"
	svcinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
	"knative.dev/pkg/tracker"

	"github.com/kelseyhightower/envconfig"
	splitterinformer "github.com/triggermesh/routing/pkg/client/generated/injection/informers/flow/v1alpha1/splitter"
	splitterreconciler "github.com/triggermesh/routing/pkg/client/generated/injection/reconciler/flow/v1alpha1/splitter"
)

type SplitterEnv struct {
	Name      string `envconfig:"SPLITTER_SERVICE" required:"true"`
	Namespace string `envconfig:"ROUTING_NAMESPACE" required:"true"`
}

// New creates a Reconciler and returns the result of NewImpl.
func New(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	var env SplitterEnv
	if err := envconfig.Process("", &env); err != nil {
		panic(fmt.Sprintf("Failed to process Splitter env: %v", err))
	}

	splitterInformer := splitterinformer.Get(ctx)
	svcInformer := svcinformer.Get(ctx)

	r := &Reconciler{
		ServiceLister: svcInformer.Lister(),
		Splitter:      env,
	}

	impl := splitterreconciler.NewImpl(ctx, r)
	r.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	logger.Info("Setting up event handlers.")

	splitterInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	svcInformer.Informer().AddEventHandler(controller.HandleAll(
		// Call the tracker's OnChanged method, but we've seen the objects
		// coming through this path missing TypeMeta, so ensure it is properly
		// populated.
		controller.EnsureTypeMeta(
			r.Tracker.OnChanged,
			corev1.SchemeGroupVersion.WithKind("Service"),
		),
	))

	return impl
}
