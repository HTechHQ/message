// and unvalid testdata for all tests to use for convenience.
//
//nolint:gochecknoglobals // allow globals in tests. The purpose of this file is to provide a wide range of valid
package message_test

import (
	"context"
	"time"

	"github.com/HTechHQ/message"
)

const (
	validTopic   = "1337"
	validMessage = "message"

	wantedSubscribers = 1000
	wantedPublishers  = 1000
)

var (
	ctx = context.Background()

	validEmptyHandlerFunc = func(ctx context.Context, e string) {}
)

type (
	simpleEvent struct{}

	eventOrTopicStructEvent struct{}

	simpleStruct    struct{}
	validSliceEvent []simpleStruct

	newUserRegisteredEvent struct {
		Username  string    `json:"userName,omitempty"`
		Email     string    `json:"email,omitempty"`
		CreatedAt time.Time `json:"createdAt,omitempty"`
		Settings  []string  `json:"settings,omitempty"`
	}

	newUserAuditLogEvent struct {
		Username  string    `json:"userName,omitempty"`
		CreatedAt time.Time `json:"createdAt,omitempty"`
		Settings  []string  `json:"settings,omitempty"`
	}

	newUserWelcomeEmailEvent struct {
		Username string `json:"userName,omitempty"`
		Email    string `json:"email,omitempty"`
	}

	shareNothingWithEvents struct {
		Name string `json:"name,omitempty"`
	}
)

func (e eventOrTopicStructEvent) Name() message.EventOrTopicName {
	return "eventOrTopicStructEvent.Name"
}

func (e validSliceEvent) Name() message.EventOrTopicName {
	return "validSliceEvent.Name"
}
