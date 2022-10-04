package main

import (
	"context"
	"fmt"

	"github.com/HTechHQ/message"
)

func main() {
	p := message.NewPubsubMem()
	ctx := context.Background()

	sub, _ := p.Subscribe(helloMessage{}, func(ctx context.Context, msg helloMessage) {
		fmt.Println(msg.Message)
	})

	p.Publish(ctx, helloMessage{"hello world!"})
	sub.Unsubscribe()
	p.Publish(ctx, helloMessage{"hello world!"})

	p.Shutdown(ctx)
	// output: hello world!
}

// helloMessage is a message passed from a publisher to potentially many subscribers.
type helloMessage struct {
	Message string
}
