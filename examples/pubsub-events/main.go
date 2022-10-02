package main

import (
	"context"
	"fmt"
	"time"

	"github.com/HTechHQ/message"
)

// newUserRegistered is the actual event that's fired after a new user got registered.
type newUserRegistered struct {
	Username  string
	Email     string
	CreatedAt time.Time
	Settings  []string
}

func main() {
	var publisher message.PubSub = message.NewPubsubMem()

	publisher.Subscribe(newUserRegistered{}, func(ctx context.Context, e newUserRegistered) {
		// log the registration of the new user
		fmt.Println(e.Username, e.Email, e.Settings, e.CreatedAt.Format("2006.01.02"))
	})
	publisher.Subscribe(newUserRegistered{}, func(ctx context.Context, e newUserRegistered) {
		fmt.Printf("Preparing welcome email for new user: %s to: %s\n", e.Username, e.Email)
	})

	publisher.Publish(context.Background(), newUserRegistered{
		Username:  "Max",
		Email:     "max@example.com",
		CreatedAt: time.Now(),
		Settings:  []string{"setting"},
	})

	publisher.Shutdown(context.Background())
	// output:
	// Max [setting] 2021.04.25
	// Preparing welcome email for new user: Max to: max@example.com
}
