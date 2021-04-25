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
	Message interface{}

	// HandlerFunc must have the signature func(arg Message) {}
	HandlerFunc interface{}
)

// PubSub is the interface providing a publish and subscribe messaging.
type PubSub interface {
	Publish(topic string, msg Message) error
	Subscribe(topic string, h HandlerFunc) (*Subscriber, error)
	Shutdown(ctx context.Context)
}
