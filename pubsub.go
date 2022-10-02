package message

import (
	"context"
	"errors"
)

var (
	ErrInvalidEvent          = errors.New("invalid EventOrMessage")
	ErrInvalidHandler        = errors.New("invalid HandlerFunc func signature")
	ErrParamConversionFailed = errors.New("invalid parameter types")
)

type (
	// EventOrTopicName is used to give an Event or a Topic a name, so publishers and subscribers can use it.
	EventOrTopicName string

	// EventOrTopic returns the EventOrTopicName of an EventOrMessage. It is optional and does not have to be
	// implemented by each EventOrMessage. If it's not implemented the struct type is used as EventOrTopicName instead.
	EventOrTopic interface {
		EventOrTopicName() EventOrTopicName
	}

	// EventOrMessage is an event or message passed between publishers and subscribers.
	// The type of EventOrMessage has to be a named struct and will be automatically converted
	// from the publisher's type to the subscriber's type, if they share a EventOrTopicName.
	EventOrMessage interface{}

	// HandlerFunc is the subscriber's handler and must have the signature:
	// func(ctx context.Context, eom EventOrMessage) {}.
	// HandlerFunc can not be more specific in this definition, so that each call to Subscribe() can give the most
	// specific parameter list, it wants to use for its domain.
	HandlerFunc interface{}
)

// PubSub is the interface providing a publish and subscribe messaging.
type PubSub interface {
	// Publish sends the EventOrMessage to all active subscribers of the given topic.
	// Publish will convert the message type to the message type of each subscriber and returns errors for all the
	// subscribers where the conversion failed.
	Publish(ctx context.Context, eom EventOrMessage) error

	// Subscribe registers h as a handler function for the given EventOrTopicName.
	// The returned Subscriber can `Unsubscribe()` from the EventOrTopicName.
	// The handler func must accept context.Context and EventOrMessage argument, and it will be converted automatically.
	// If EventOrMessage implements the EventOrTopic interface, it is called to determine the EventOrTopicName,
	// otherwise the struct name is used.
	Subscribe(eom EventOrMessage, h HandlerFunc) (*Subscriber, error)

	// Shutdown will block and wait for all handlers to finish or for the context to time out, whichever happens first.
	Shutdown(ctx context.Context)
}
