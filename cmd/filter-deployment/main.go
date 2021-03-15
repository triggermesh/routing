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
	"log"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	configmap "knative.dev/pkg/configmap/informer"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/system"

	"github.com/triggermesh/filter/pkg/reconciler/config"
	"github.com/triggermesh/filter/pkg/reconciler/filter"

	"knative.dev/pkg/injection/sharedmain"

	filterv1alpha1 "github.com/triggermesh/filter/pkg/client/generated/clientset/internalclientset"
	filterinformers "github.com/triggermesh/filter/pkg/client/generated/informers/externalversions"
)

const (
	defaultMetricsPort = 9092
	component          = "filter-service"
)

type envConfig struct {
	Namespace string `envconfig:"NAMESPACE" required:"true"`
	// TODO: change this environment variable to something like "PodGroupName".
	PodName       string `envconfig:"POD_NAME" required:"true"`
	ContainerName string `envconfig:"CONTAINER_NAME" required:"true"`
	Port          int    `envconfig:"FILTER_PORT" default:"8080"`
}

func main() {
	ctx := signals.NewContext()

	// Report stats on Go memory usage every 30 seconds.
	metrics.MemStatsOrDie(ctx)

	cfg := sharedmain.ParseAndGetConfigOrDie()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env var", zap.Error(err))
	}

	ctx, _ = injection.Default.SetupInformers(ctx, cfg)
	kubeClient := kubeclient.Get(ctx)

	loggingConfig, err := config.GetLoggingConfig(ctx, system.Namespace(), logging.ConfigMapName())
	if err != nil {
		log.Fatal("Error loading/parsing logging configuration:", err)
	}
	sl, atomicLevel := logging.NewLoggerFromConfig(loggingConfig, component)
	logger := sl.Desugar()
	defer flush(sl)

	logger.Info("Starting the Filter")

	filterClient := filterv1alpha1.NewForConfigOrDie(cfg)
	filterFactory := filterinformers.NewSharedInformerFactory(filterClient, controller.GetResyncPeriod(ctx))
	filterInformer := filterFactory.Routing().V1alpha1().Filters()

	// Watch the logging config map and dynamically update logging levels.
	configMapWatcher := configmap.NewInformedWatcher(kubeClient, system.Namespace())
	// Watch the observability config map and dynamically update metrics exporter.
	updateFunc, err := metrics.UpdateExporterFromConfigMapWithOpts(ctx, metrics.ExporterOptions{
		Component:      component,
		PrometheusPort: defaultMetricsPort,
	}, sl)
	if err != nil {
		logger.Fatal("Failed to create metrics exporter update function", zap.Error(err))
	}
	configMapWatcher.Watch(metrics.ConfigMapName(), updateFunc)
	// TODO change the component name to broker once Stackdriver metrics are approved.
	// Watch the observability config map and dynamically update request logs.
	configMapWatcher.Watch(logging.ConfigMapName(), logging.UpdateLevelFromConfigMap(sl, atomicLevel, component))

	// reporter := filter.NewStatsReporter(env.ContainerName, kmeta.ChildName(env.PodName, uuid.New().String()))

	// We are running both the receiver (takes messages in from the Broker) and the dispatcher (send
	// the messages to the triggers' subscribers) in this binary.
	handler, err := filter.NewHandler(logger, filterInformer.Lister(), env.Port)
	if err != nil {
		logger.Fatal("Error creating Handler", zap.Error(err))
	}

	// configMapWatcher does not block, so start it first.
	if err = configMapWatcher.Start(ctx.Done()); err != nil {
		logger.Warn("Failed to start ConfigMap watcher", zap.Error(err))
	}

	// Start all of the informers and wait for them to sync.
	logger.Info("Starting informer.")

	go filterFactory.Start(ctx.Done())
	filterFactory.WaitForCacheSync(ctx.Done())

	// Start blocks forever.
	logger.Info("Filter starting...")

	err = handler.Start(ctx)
	if err != nil {
		logger.Fatal("handler.Start() returned an error", zap.Error(err))
	}
	logger.Info("Exiting...")
}

func flush(logger *zap.SugaredLogger) {
	_ = logger.Sync()
	metrics.FlushExporter()
}