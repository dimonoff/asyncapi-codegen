package generators

// Side represents which role an application plays in the AsyncAPI specification —
// either the component that publishes messages to a channel, or the component
// that subscribes to (consumes) messages from a channel.
type Side string

const (
	// SideIsPublisher represents the component that publishes messages to channels.
	// It maps to the "application" / send side of the AsyncAPI specification.
	SideIsPublisher Side = "publisher"
	// SideIsSubscriber represents the component that subscribes to (receives) messages
	// from channels. It maps to the "user" / receive side of the AsyncAPI specification.
	SideIsSubscriber Side = "subscriber"
)
