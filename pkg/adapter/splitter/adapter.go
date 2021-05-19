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
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"

	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/eventing/pkg/utils"

	routinglisters "github.com/triggermesh/routing/pkg/client/generated/listers/flow/v1alpha1"
)

const (
	// TODO make these constants configurable (either as env variables, config map, or part of broker spec).
	//  Issue: https://github.com/knative/eventing/issues/1777
	// Constants for the underlying HTTP Client transport. These would enable better connection reuse.
	// Set them on a 10:1 ratio, but this would actually depend on the Triggers' subscribers and the workload itself.
	// These are magic numbers, partly set based on empirical evidence running performance workloads, and partly
	// based on what serving is doing. See https://github.com/knative/serving/blob/master/pkg/network/transports.go.
	defaultMaxIdleConnections        = 1000
	defaultMaxIdleConnectionsPerHost = 100
)

type splitterRef struct {
	name      string
	namespace string
}

// Handler parses Cloud Events, determines if they pass a filter, and sends them to a subscriber.
type Handler struct {
	// receiver receives incoming HTTP requests
	receiver *kncloudevents.HTTPMessageReceiver
	// sender sends requests to downstream services
	sender *kncloudevents.HTTPMessageSender
	// reporter reports stats of status code and dispatch time
	// reporter StatsReporter

	splitterLister routinglisters.SplitterLister
	logger         *zap.Logger
}

// NewHandler creates a new Handler and its associated MessageReceiver. The caller is responsible for
// Start()ing the returned Handler.
func NewHandler(logger *zap.Logger, splitterLister routinglisters.SplitterLister, port int) (*Handler, error) {
	kncloudevents.ConfigureConnectionArgs(&kncloudevents.ConnectionArgs{
		MaxIdleConns:        defaultMaxIdleConnections,
		MaxIdleConnsPerHost: defaultMaxIdleConnectionsPerHost,
	})

	sender, err := kncloudevents.NewHTTPMessageSenderWithTarget("")
	if err != nil {
		return nil, fmt.Errorf("failed to create message sender: %w", err)
	}

	return &Handler{
		receiver:       kncloudevents.NewHTTPMessageReceiver(port),
		sender:         sender,
		splitterLister: splitterLister,
		logger:         logger,
	}, nil
}

// Start begins to receive messages for the handler.
//
// HTTP POST requests to the root path (/) are accepted.
//
// This method will block until ctx is done.
func (h *Handler) Start(ctx context.Context) error {
	return h.receiver.StartListen(ctx, h)
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ref, err := parseRequestURI(request.RequestURI)
	if err != nil {
		h.logger.Info("Unable to parse path as splitter", zap.Error(err), zap.String("path", request.RequestURI))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := request.Context()

	message := cehttp.NewMessageFromHttpRequest(request)
	// cannot be err, but makes linter complain about missing err check
	//nolint
	defer message.Finish(nil)

	event, err := binding.ToEvent(ctx, message)
	if err != nil {
		h.logger.Warn("failed to extract event from request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Debug("Received message", zap.Any("splitterRef", ref))

	s, err := h.splitterLister.Splitters(ref.namespace).Get(ref.name)
	if err != nil {
		h.logger.Info("Unable to get the Splitter", zap.Error(err), zap.Any("splitterRef", ref))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	for i, e := range h.split(s.Spec.Path, event) {
		e.SetID(fmt.Sprintf("%s-%d", event.ID(), i))
		e.SetType(s.Spec.CEContext.Type)
		e.SetSource(s.Spec.CEContext.Source)
		for key, value := range s.Spec.CEContext.Extensions {
			e.SetExtension(key, value)
		}
		// we may want to keep responses and send them back to the source
		_, err := h.sendEvent(ctx, request.Header, s.Status.SinkURI.String(), e)
		if err != nil {
			h.logger.Error("failed to send the event", zap.Error(err))
		}
	}

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) split(path string, e *event.Event) []*event.Event {
	var result []*event.Event

	val := gjson.Get(string(e.Data()), path)
	if !val.IsArray() {
		return result
	}
	for _, v := range val.Array() {
		newCE := cloudevents.NewEvent()
		newCE.SetData(cloudevents.ApplicationJSON, v.Raw)
		result = append(result, &newCE)
	}
	return result
}

func (h *Handler) sendEvent(ctx context.Context, headers http.Header, target string, event *cloudevents.Event) (*http.Response, error) {
	// Send the event to the subscriber
	req, err := h.sender.NewCloudEventRequestWithTarget(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("failed to create the request: %w", err)
	}

	message := binding.ToMessage(event)
	// cannot be err, but makes linter complain about missing err check
	//nolint
	defer message.Finish(nil)

	additionalHeaders := utils.PassThroughHeaders(headers)
	err = kncloudevents.WriteHTTPRequestWithAdditionalHeaders(ctx, message, req, additionalHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	resp, err := h.sender.Send(req)
	if err != nil {
		err = fmt.Errorf("failed to dispatch message: %w", err)
	}

	return resp, err
}

func parseRequestURI(path string) (splitterRef, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return splitterRef{}, fmt.Errorf("incorrect number of parts in the path, expected 4, actual %d, '%s'", len(parts), path)
	}
	if parts[0] != "" {
		return splitterRef{}, fmt.Errorf("text before the first slash, actual '%s'", path)
	}
	if parts[1] != "splitters" {
		return splitterRef{}, fmt.Errorf("incorrect prefix, expected 'splitters', actual '%s'", path)
	}
	return splitterRef{
		namespace: parts[2],
		name:      parts[3],
	}, nil
}
