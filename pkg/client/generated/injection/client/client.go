/*
Copyright (c) 2021 TriggerMesh Inc.

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

// Code generated by injection-gen. DO NOT EDIT.

package client

import (
	context "context"

	internalclientset "github.com/triggermesh/routing/pkg/client/generated/clientset/internalclientset"
	rest "k8s.io/client-go/rest"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterClient(withClient)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, Key{}, internalclientset.NewForConfigOrDie(cfg))
}

// Get extracts the internalclientset.Interface client from the context.
func Get(ctx context.Context) internalclientset.Interface {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		if injection.GetConfig(ctx) == nil {
			logging.FromContext(ctx).Panic(
				"Unable to fetch github.com/triggermesh/routing/pkg/client/generated/clientset/internalclientset.Interface from context. This context is not the application context (which is typically given to constructors via sharedmain).")
		} else {
			logging.FromContext(ctx).Panic(
				"Unable to fetch github.com/triggermesh/routing/pkg/client/generated/clientset/internalclientset.Interface from context.")
		}
	}
	return untyped.(internalclientset.Interface)
}
