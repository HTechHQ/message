// and unvalid testdata for all tests to use for convenience.
//
//nolint:gochecknoglobals // allow globals in tests. The purpose of this file is to provide a wide range of valid
package message_test

import (
	"context"

	"github.com/HTechHQ/message"
)

const (
	validTopic   = "1337"
	validMessage = "message"
)

var (
	validCtx = context.Background()

	validEmptyHandlerFunc = func(ctx context.Context, e string) {}
)

type (
	simpleStruct            struct{}
	simpleEvent             struct{}
	eventOrTopicStructEvent struct{}
	validSliceEvent         []simpleStruct
)

func (e eventOrTopicStructEvent) Name() message.EventOrTopicName {
	return "eventOrTopicStructEvent.Name"
}

func (e validSliceEvent) Name() message.EventOrTopicName {
	return "validSliceEvent.Name"
}
