//go:generate go run ../../../../cmd/asyncapi-codegen -g types,subscriber -p none -i ./asyncapi.yaml -o ./none/asyncapi.gen.go
//go:generate go run ../../../../cmd/asyncapi-codegen -g types,subscriber -p snake --convert-keys snake -i ./asyncapi.yaml -o ./snake/asyncapi.gen.go
//go:generate go run ../../../../cmd/asyncapi-codegen -g types,subscriber -p camel --convert-keys camel -i ./asyncapi.yaml -o ./camel/asyncapi.gen.go
//go:generate go run ../../../../cmd/asyncapi-codegen -g types,subscriber -p kebab --convert-keys kebab -i ./asyncapi.yaml -o ./kebab/asyncapi.gen.go

package issue129

import (
	"context"
	"testing"

	"github.com/dimonoff/asyncapi-codegen/pkg/extensions"
	"github.com/dimonoff/asyncapi-codegen/pkg/extensions/middlewares"
	"github.com/dimonoff/asyncapi-codegen/pkg/utils"
	testutil "github.com/dimonoff/asyncapi-codegen/test"
	"github.com/dimonoff/asyncapi-codegen/test/v3/issues/129/camel"
	"github.com/dimonoff/asyncapi-codegen/test/v3/issues/129/kebab"
	"github.com/dimonoff/asyncapi-codegen/test/v3/issues/129/none"
	"github.com/dimonoff/asyncapi-codegen/test/v3/issues/129/snake"
	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	brokers, cleanup := testutil.BrokerControllers(t)
	defer cleanup()

	for _, b := range brokers {
		suite.Run(t, NewSuite(b))
	}
}

type Suite struct {
	broker extensions.BrokerController

	suite.Suite
}

func NewSuite(broker extensions.BrokerController) *Suite {
	return &Suite{
		broker: broker,
	}
}

func (suite *Suite) TestWithNoneKeyConversion() {
	// Create a channel to intercept message before sending to broker and after
	// reception from broker
	interceptor := make(chan extensions.BrokerMessage, 8)
	defer close(interceptor)

	// Create none user
	user, err := none.NewSubscriberController(suite.broker, none.WithMiddlewares(middlewares.Intercepter(interceptor)))
	suite.Require().NoError(err)
	defer user.Close(context.Background())

	// Send the message
	err = user.SendToReceiveTestOperation(context.Background(), none.TestMessageFromTestChannel{
		Payload: none.TestSchema{
			ThisIsAProperty: utils.ToPointer("value"),
		},
	})
	suite.Require().NoError(err)

	// Check sent message to broker
	bMsg := <-interceptor

	// Check that the additional properties are at the level 0 of payload
	suite.Require().Equal("{\"This_is a-Property\":\"value\"}", string(bMsg.Payload))
}

func (suite *Suite) TestWithSnakeKeyConversion() {
	// Create a channel to intercept message before sending to broker and after
	// reception from broker
	interceptor := make(chan extensions.BrokerMessage, 8)
	defer close(interceptor)

	// Create snake user
	user, err := snake.NewSubscriberController(suite.broker, snake.WithMiddlewares(middlewares.Intercepter(interceptor)))
	suite.Require().NoError(err)
	defer user.Close(context.Background())

	// Send the message
	err = user.SendToReceiveTestOperation(context.Background(), snake.TestMessageFromTestChannel{
		Payload: snake.TestSchema{
			ThisIsAProperty: utils.ToPointer("value"),
		},
	})
	suite.Require().NoError(err)

	// Check sent message to broker
	bMsg := <-interceptor

	// Check that the additional properties are at the level 0 of payload
	suite.Require().Equal("{\"this_is_a_property\":\"value\"}", string(bMsg.Payload))
}

func (suite *Suite) TestWithKebabKeyConversion() {
	// Create a channel to intercept message before sending to broker and after
	// reception from broker
	interceptor := make(chan extensions.BrokerMessage, 8)
	defer close(interceptor)

	// Create kebab user
	user, err := kebab.NewSubscriberController(suite.broker, kebab.WithMiddlewares(middlewares.Intercepter(interceptor)))
	suite.Require().NoError(err)
	defer user.Close(context.Background())

	// Send the message
	err = user.SendToReceiveTestOperation(context.Background(), kebab.TestMessageFromTestChannel{
		Payload: kebab.TestSchema{
			ThisIsAProperty: utils.ToPointer("value"),
		},
	})
	suite.Require().NoError(err)

	// Check sent message to broker
	bMsg := <-interceptor

	// Check that the additional properties are at the level 0 of payload
	suite.Require().Equal("{\"this-is-a-property\":\"value\"}", string(bMsg.Payload))
}

func (suite *Suite) TestWithCamelKeyConversion() {
	// Create a channel to intercept message before sending to broker and after
	// reception from broker
	interceptor := make(chan extensions.BrokerMessage, 8)
	defer close(interceptor)

	// Create camel user
	user, err := camel.NewSubscriberController(suite.broker, camel.WithMiddlewares(middlewares.Intercepter(interceptor)))
	suite.Require().NoError(err)
	defer user.Close(context.Background())

	// Send the message
	err = user.SendToReceiveTestOperation(context.Background(), camel.TestMessageFromTestChannel{
		Payload: camel.TestSchema{
			ThisIsAProperty: utils.ToPointer("value"),
		},
	})
	suite.Require().NoError(err)

	// Check sent message to broker
	bMsg := <-interceptor

	// Check that the additional properties are at the level 0 of payload
	suite.Require().Equal("{\"ThisIsAProperty\":\"value\"}", string(bMsg.Payload))
}
