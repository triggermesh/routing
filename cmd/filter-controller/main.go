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

package main

import (
	// The set of controllers this controller process runs.
	"log"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	// This defines the shared main for injected controllers.
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"

	"github.com/triggermesh/filter/pkg/reconciler/config"
	"github.com/triggermesh/filter/pkg/reconciler/controller"
)

const (
	component = "filter-service"
)

func main() {
	var filterEnv controller.FilterService
	if err := envconfig.Process("", &filterEnv); err != nil {
		log.Fatal("Failed to process env var", zap.Error(err))
	}

	ctx := signals.NewContext()
	sharedmain.MainWithContext(config.WithFilterService(ctx, filterEnv),
		component, controller.New)
}
