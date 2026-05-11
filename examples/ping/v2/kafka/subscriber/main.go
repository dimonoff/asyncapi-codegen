//go:generate go run ../../../../../cmd/asyncapi-codegen -g subscriber,types -p main -i ../../asyncapi.yaml -o ./subscriber.gen.go

package main

import (
	"context"

	"github.com/dimonoff/asyncapi-codegen/pkg/extensions/brokers/kafka"
	"github.com/dimonoff/asyncapi-codegen/pkg/extensions/loggers"
	"github.com/dimonoff/asyncapi-codegen/pkg/extensions/middlewares"
	testutil "github.com/dimonoff/asyncapi-codegen/pkg/utils/test"
)

func main() {
	// Get broker address based on the environment, it will returns an address like "kafka:9092"
	// Note: this is not needed in your application, you can directly use the address
	addr := testutil.BrokerAddress(testutil.BrokerAddressParams{
		DockerizedAddr: "kafka",
		Port:           "9092",
	})

	// Instantiate a Kafka controller with a logger
	logger := loggers.NewText()
	broker, err := kafka.NewController(
		[]string{addr},                  // List of hosts
		kafka.WithLogger(logger),        // Attach an internal logger
		kafka.WithGroupID("ping-users"), // Change group id
	)
	if err != nil {
		panic(err)
	}

	// Create a new client controller
	ctrl, err := NewSubscriberController(
		broker,             // Attach the kafka controller
		WithLogger(logger), // Attach an internal logger
		WithMiddlewares(middlewares.Logging(logger))) // Attach a middleware to log messages
	if err != nil {
		panic(err)
	}
	defer ctrl.Close(context.Background())

	// Make a new ping message
	req := NewPingMessage()
	req.Payload = "ping"

	// Create the publication function to send the message
	// Note: it will indefinitely wait to publish as context has no timeout
	publicationFunc := func(ctx context.Context) error {
		return ctrl.PublishPing(ctx, req)
	}

	// The following function will subscribe to the 'pong' channel, execute the publication
	// function and wait for a response. The response will be detected through its
	// correlation ID.
	//
	// Note: it will indefinitely wait for messages as context has no timeout
	_, err = ctrl.WaitForPong(context.Background(), &req, publicationFunc)
	if err != nil {
		panic(err)
	}
}
