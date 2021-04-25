package message

import (
	"context"
	"errors"
)

var (
	ErrInvalidHandler        = errors.New("invalid HandlerFunc func signature")
	ErrParamConversionFailed = errors.New("invalid parameter types")
)

type (
	// Message is a message passed between publishers and subscribers.
	// The message will be automatically converted from the publisher's type to the subscriber's type.
	Message interface{}

	// HandlerFunc is the subscribers handler and must have the signature: func(arg Message) {}
	HandlerFunc interface{}
)

// PubSub is the interface providing a publish and subscribe messaging.
type PubSub interface {
	// Publish sends the msg to all active subscribers of the given topic.
	// Publish will convert the message type to the message type of each subscriber and returns errors for all the
	// subscribers where the conversion failed.
	Publish(topic string, msg Message) error

	// Subscribe registers h as a handler function for the given topic.
	// The returned Subscriber can `Unsubscribe()` from the topic.
	// The handler func must accept one Message argument and it will be converted automatically.
	Subscribe(topic string, h HandlerFunc) (*Subscriber, error)

	// Shutdown will block and wait for all handlers to finish or for the context to time out, whichever happens first.
	Shutdown(ctx context.Context)
}
