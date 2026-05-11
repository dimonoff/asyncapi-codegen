package extensions

import (
	"errors"
	"fmt"
)

var (
	// ErrAsyncAPI is the generic error for AsyncAPI generated code.
	ErrAsyncAPI = errors.New("error when using AsyncAPI")

	// ErrContextCanceled is given when a given context is canceled.
	ErrContextCanceled = fmt.Errorf("%w: context canceled", ErrAsyncAPI)

	// ErrNilBrokerController is raised when a nil broker controller is used.
	ErrNilBrokerController = fmt.Errorf("%w: nil broker controller has been used", ErrAsyncAPI)

	// ErrNilPublisherHandler is raised when a nil handler is passed to a PublisherController.
	ErrNilPublisherHandler = fmt.Errorf("%w: nil publisher handler has been used", ErrAsyncAPI)

	// ErrNilSubscriberHandler is raised when a nil handler is passed to a SubscriberController.
	ErrNilSubscriberHandler = fmt.Errorf("%w: nil subscriber handler has been used", ErrAsyncAPI)

	// ErrAlreadySubscribedChannel is raised when a subscription is done twice
	// or more without unsubscribing.
	ErrAlreadySubscribedChannel = fmt.Errorf("%w: the channel has already been subscribed", ErrAsyncAPI)

	// ErrSubscriptionCanceled is raised when expecting something and the subscription has been canceled before it happens.
	ErrSubscriptionCanceled = fmt.Errorf("%w: the subscription has been canceled", ErrAsyncAPI)

	// ErrNoCorrelationIDSet is raised when a correlation ID is expected, but none is detected.
	ErrNoCorrelationIDSet = fmt.Errorf("%w: no correlation ID but one is expected", ErrAsyncAPI)

	// ErrChannelAddressEmpty is raised when a given channel address is empty,
	// when dynamically set from message.
	ErrChannelAddressEmpty = fmt.Errorf("%w: channel address empty", ErrAsyncAPI)
)
