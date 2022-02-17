package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	cloudeventsotel "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cloudeventshttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/pavolloffay/knative-tracing/httplogging"
	"github.com/pavolloffay/knative-tracing/tracing"
)

const ENV_BROKER_URL = "ENV_BROKER_URL"

var tracer = otel.Tracer("first-in-app-tracer")

func main() {
	log.Print("First application started.")

	// Use instrumented cloud event client - it will inject trace context into outgoing requests (send event).
	c, err := cloudeventsotel.NewClientHTTP([]cloudeventshttp.Option{}, []client.Option{})
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	_, err = tracing.InitOTEL(context.Background())
	if err != nil {
		return
	}

	eventHandler := &handler{client: c}
	// Wrap application handler into OTEL to create HTTP server spans and propagate context into the application.
	// The OTEL handler extracts span context from incoming request, creates span and injects span context into go context.
	// The application handler can get span context and link application span with trace started by knative.
	otelHandler := otelhttp.NewHandler(eventHandler, "/")
	rootHandler := &httplogging.LoggingHandler{Wrapped: otelHandler}
	http.ListenAndServe(":8090", rootHandler)
}

type handler struct {
	client cloudevents.Client
}

var _ http.Handler = (*handler)(nil)

func (s *handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	ctx, span := tracer.Start(req.Context(), "user_handler: /")
	defer span.End()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	span.SetAttributes(attribute.String("request-body", string(body)))

	event := cloudevents.NewEvent()
	event.SetSource("github/com/pavolloffay")
	event.SetType("httpbody")
	event.SetID(uuid.New().String())
	event.SetData(cloudevents.TextPlain, fmt.Sprintf("hello from first, traceid=%s", span.SpanContext().TraceID()))
	ctx = cloudevents.ContextWithTarget(ctx, os.Getenv(ENV_BROKER_URL))

	span.SetAttributes(attribute.String("event-id", event.ID()))

	result := s.client.Send(ctx, event)
	if protocol.IsNACK(result) {
		log.Printf("send error: %v\n", result)
		writer.Write([]byte("Failed to send cloudevent: " + result.Error()))
		writer.WriteHeader(500)
	} else {
		writer.Write([]byte(fmt.Sprintf("Hello, traceid=%s\n\n", span.SpanContext().TraceID())))
		writer.WriteHeader(200)
	}
}
