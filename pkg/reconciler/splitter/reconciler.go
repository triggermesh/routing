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

	"knative.dev/pkg/reconciler"

	"github.com/triggermesh/routing/pkg/apis/flow/v1alpha1"
	routingv1alpha1 "github.com/triggermesh/routing/pkg/apis/flow/v1alpha1"
	splitterreconciler "github.com/triggermesh/routing/pkg/client/generated/injection/reconciler/flow/v1alpha1/splitter"
	listersv1alpha1 "github.com/triggermesh/routing/pkg/client/generated/listers/flow/v1alpha1"
	"github.com/triggermesh/routing/pkg/reconciler/common"
)

// Reconciler implements addressableservicereconciler.Interface for
// AddressableService resources.
type Reconciler struct {
	base           common.GenericServiceReconciler
	splitterLister func(namespace string) listersv1alpha1.SplitterNamespaceLister
	adapterCfg     *adapterConfig
}

// Check that our Reconciler implements Interface
var _ splitterreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *routingv1alpha1.Splitter) reconciler.Event {
	// inject source into context for usage in reconciliation logic
	ctx = v1alpha1.WithRouter(ctx, o)

	return r.base.ReconcileAdapter(ctx, r)
}
