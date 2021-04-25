package main

import (
	"context"
	"fmt"

	"github.com/HTechHQ/message"
)

func main() {
	c := message.NewPubsubMem()

	sub, _ := c.Subscribe("topic-name", func(msg []byte) {
		fmt.Println(string(msg))
	})

	c.Publish("topic-name", []byte("hello world!"))
	sub.Unsubscribe()
	c.Publish("topic-name", []byte("hello world!"))

	c.Shutdown(context.TODO())
	// output: hello world!
}
