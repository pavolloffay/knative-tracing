package main

import (
	"context"
	"log"

	cloudeventsotel "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/pavolloffay/knative-tracing/tracing"
)

var tracer = otel.Tracer("second-in-app-tracer")

func main() {
	log.Print("Second application starting.")

	_, err := tracing.InitOTEL(context.Background())

	// Create instrumented cloudevents client.
	// In this case the client is used as server to receive events
	// The instrumentation ensures trace context propagation and creates spans.
	c, err := cloudeventsotel.NewClientHTTP([]cloudeventshttp.Option{}, []client.Option{})
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), receiveEvent))
}

func receiveEvent(ctx context.Context, event cloudevents.Event) cloudevents.Result {
	ctx, span := tracer.Start(ctx, "user_function: receiveEvent")
	defer span.End()

	log.Printf("Event received: %s\n", event)
	log.Printf("Message from received event: %s", string(event.Data()))

	span.SetAttributes(attribute.String("event.data.msg", string(event.Data())))
	return nil
}
