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

package filter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"

	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/eventing/pkg/utils"

	routingv1alpha1 "github.com/triggermesh/routing/pkg/apis/flow/v1alpha1"
	routinglisters "github.com/triggermesh/routing/pkg/client/generated/listers/flow/v1alpha1"
	"github.com/triggermesh/routing/pkg/eventfilter"
	"github.com/triggermesh/routing/pkg/eventfilter/cel"
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

type filterRef struct {
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

	filterLister routinglisters.FilterLister
	logger       *zap.Logger

	// expressions is the map of trigger refs with precompiled CEL expressions
	// TODO (tzununbekov): Add cleanup
	expressions *expressionStorage
}

// NewHandler creates a new Handler and its associated MessageReceiver. The caller is responsible for
// Start()ing the returned Handler.
func NewHandler(logger *zap.Logger, filterLister routinglisters.FilterLister, port int) (*Handler, error) {
	kncloudevents.ConfigureConnectionArgs(&kncloudevents.ConnectionArgs{
		MaxIdleConns:        defaultMaxIdleConnections,
		MaxIdleConnsPerHost: defaultMaxIdleConnectionsPerHost,
	})

	sender, err := kncloudevents.NewHTTPMessageSenderWithTarget("")
	if err != nil {
		return nil, fmt.Errorf("failed to create message sender: %w", err)
	}

	return &Handler{
		receiver:     kncloudevents.NewHTTPMessageReceiver(port),
		sender:       sender,
		filterLister: filterLister,
		logger:       logger,
		expressions:  newExpressionStorage(),
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
		h.logger.Info("Unable to parse path as filter", zap.Error(err), zap.String("path", request.RequestURI))
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

	h.logger.Debug("Received message", zap.Any("filterRef", ref))

	f, err := h.getFilter(ref)
	if err != nil {
		h.logger.Info("Unable to get the Filter", zap.Error(err), zap.Any("filterRef", ref))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	cond, exists := h.expressions.get(f.UID, f.Generation)
	if !exists {
		cond, err = cel.CompileExpression(f.Spec.Expression)
		if err != nil {
			h.logger.Info("Failed to compile filter expression", zap.Error(err), zap.Any("filterRef", ref))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		h.expressions.set(f.UID, f.Generation, cond)
	}

	filterResult := filterEvent(ctx, cond, *event)
	if filterResult == eventfilter.FailFilter {
		return
	}

	event = updateAttributes(f.Status, event)
	h.send(ctx, writer, request.Header, f.Status.SinkURI.String(), event)
}

func updateAttributes(fs routingv1alpha1.FilterStatus, event *event.Event) *event.Event {
	if len(fs.CloudEventAttributes) == 1 {
		event.SetType(fs.CloudEventAttributes[0].Type)
		event.SetSource(fs.CloudEventAttributes[0].Source)
	}
	return event
}

func (h *Handler) send(ctx context.Context, writer http.ResponseWriter, headers http.Header, target string, event *cloudevents.Event) {
	// send the event to trigger's subscriber
	response, err := h.sendEvent(ctx, headers, target, event)
	if err != nil {
		h.logger.Error("failed to send event", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Successfully dispatched message", zap.Any("target", target))

	// If there is an event in the response write it to the response
	_, err = h.writeResponse(ctx, writer, response, target)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
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

// The return values are the status
func (h *Handler) writeResponse(ctx context.Context, writer http.ResponseWriter, resp *http.Response, target string) (int, error) {
	response := cehttp.NewMessageFromHttpResponse(resp)
	// cannot be err, but makes linter complain about missing err check
	//nolint
	defer response.Finish(nil)

	if response.ReadEncoding() == binding.EncodingUnknown {
		// Response doesn't have a ce-specversion header nor a content-type matching a cloudevent event format
		// Just read a byte out of the reader to see if it's non-empty, we don't care what it is,
		// just that it is not empty. This means there was a response and it's not valid, so treat
		// as delivery failure.
		body := make([]byte, 1)
		n, _ := response.BodyReader.Read(body)
		response.BodyReader.Close()
		if n != 0 {
			// Note that we could just use StatusInternalServerError, but to distinguish
			// between the failure cases, we use a different code here.
			writer.WriteHeader(http.StatusBadGateway)
			return http.StatusBadGateway, errors.New("received a non-empty response not recognized as CloudEvent. The response MUST be or empty or a valid CloudEvent")
		}
		h.logger.Debug("Response doesn't contain a CloudEvent, replying with an empty response", zap.Any("target", target))
		writer.WriteHeader(resp.StatusCode)
		return resp.StatusCode, nil
	}

	event, err := binding.ToEvent(ctx, response)
	if err != nil {
		// Like in the above case, we could just use StatusInternalServerError, but to distinguish
		// between the failure cases, we use a different code here.
		writer.WriteHeader(http.StatusBadGateway)
		// Malformed event, reply with err
		return http.StatusBadGateway, err
	}

	eventResponse := binding.ToMessage(event)
	// cannot be err, but makes linter complain about missing err check
	//nolint
	defer eventResponse.Finish(nil)

	if err := cehttp.WriteResponseWriter(ctx, eventResponse, resp.StatusCode, writer); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to write response event: %w", err)
	}

	h.logger.Debug("Replied with a CloudEvent response", zap.Any("target", target))

	return resp.StatusCode, nil
}

func (h *Handler) getFilter(ref filterRef) (*routingv1alpha1.Filter, error) {
	return h.filterLister.Filters(ref.namespace).Get(ref.name)
}

func filterEvent(ctx context.Context, filter cel.ConditionalFilter, event cloudevents.Event) eventfilter.FilterResult {
	var filters eventfilter.Filters
	if filter.Expression != nil {
		filters = append(filters, &filter)
	}

	return filters.Filter(ctx, event)
}

func parseRequestURI(path string) (filterRef, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return filterRef{}, fmt.Errorf("incorrect number of parts in the path, expected 4, actual %d, '%s'", len(parts), path)
	}
	if parts[0] != "" {
		return filterRef{}, fmt.Errorf("text before the first slash, actual '%s'", path)
	}
	if parts[1] != "filters" {
		return filterRef{}, fmt.Errorf("incorrect prefix, expected 'filters', actual '%s'", path)
	}
	return filterRef{
		namespace: parts[2],
		name:      parts[3],
	}, nil
}
