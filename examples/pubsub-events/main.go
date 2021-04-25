package main

import (
	"context"
	"fmt"
	"time"

	"github.com/HTechHQ/message"
)

type newUserRegisteredEvent struct {
	Username  string
	Email     string
	CreatedAt time.Time
	Settings  []string
}

type newUserAuditLogEvent struct {
	Username  string
	CreatedAt time.Time
	Settings  []string
}

type newUserWelcomeEmailEvent struct {
	Name  string `json:"UserName"`
	Email string
}

func main() {
	c := message.NewPubsubMem()
	topic := "user-registration"

	c.Subscribe(topic, func(e newUserAuditLogEvent) {
		fmt.Println(e.Username, e.Settings, e.CreatedAt.Format("2006.01.02"))
	})
	c.Subscribe(topic, func(e newUserWelcomeEmailEvent) {
		fmt.Println(e.Name, e.Email)
	})

	c.Publish(topic, newUserRegisteredEvent{
		Username:  "Max",
		Email:     "max@example.com",
		CreatedAt: time.Now(),
		Settings:  []string{"setting"},
	})

	c.Shutdown(context.TODO())
	// output:
	// Max [setting] 2021.04.25
	// Max max@example.com
}
