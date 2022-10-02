// Valid and invalid testdata for all tests to use for convenience.
//
//nolint:gochecknoglobals // allow globals in tests. The purpose of this file is to provide a wide range of valid
package message_test

import (
	"context"
	"time"

	"github.com/HTechHQ/message"
)

const (
	wantedSubscribers = 1000
	wantedPublishers  = 1000

	sharedTopicName = "shared.topic.name"
)

var (
	ctx                   = context.Background()
	validEmptyHandlerFunc = func(ctx context.Context, e simpleEvent) {}

	ctxKey = struct{}{}
	ctxVal = "some-value"
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

func (e eventOrTopicStructEvent) EventOrTopicName() message.EventOrTopicName {
	return "eventOrTopicStructEvent.EventOrTopicName"
}

func (e validSliceEvent) EventOrTopicName() message.EventOrTopicName {
	return "validSliceEvent.EventOrTopicName"
}

func (e newUserRegisteredEvent) EventOrTopicName() message.EventOrTopicName {
	return sharedTopicName
}

func (e newUserAuditLogEvent) EventOrTopicName() message.EventOrTopicName {
	return sharedTopicName
}

func (e newUserWelcomeEmailEvent) EventOrTopicName() message.EventOrTopicName {
	return sharedTopicName
}

func (e shareNothingWithEvents) EventOrTopicName() message.EventOrTopicName {
	return sharedTopicName
}
