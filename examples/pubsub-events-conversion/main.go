package main

import (
	"context"
	"fmt"
	"time"

	"github.com/HTechHQ/message"
)

func main() {
	p := message.NewPubsubMem()
	ctx := context.Background()

	p.Subscribe(newUserAuditLogEvent{}, func(e newUserAuditLogEvent) {
		fmt.Println(e.Username, e.Settings, e.CreatedAt.Format("2006.01.02"))
	})
	p.Subscribe(newUserWelcomeEmailEvent{}, func(e newUserWelcomeEmailEvent) {
		fmt.Println(e.Name, e.Email)
	})

	p.Publish(ctx, newUserRegisteredEvent{
		Username:  "Max",
		Email:     "max@example.com",
		CreatedAt: time.Now(),
		Settings:  []string{"setting"},
	})

	p.Shutdown(context.TODO())
	// output:
	// Max [setting] 2021.04.25
	// Max max@example.com
}

type newUserRegisteredEvent struct {
	Username  string
	Email     string
	CreatedAt time.Time
	Settings  []string
}

func (e newUserRegisteredEvent) EventOrTopicName() string {
	return "new.user"
}

type newUserAuditLogEvent struct {
	Username  string
	CreatedAt time.Time
	Settings  []string
}

func (e newUserAuditLogEvent) EventOrTopicName() string {
	return "new.user"
}

type newUserWelcomeEmailEvent struct {
	Name  string `json:"UserName"`
	Email string
}

func (e newUserWelcomeEmailEvent) EventOrTopicName() string {
	return "new.user"
}
