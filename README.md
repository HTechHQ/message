[![Test Build][github-action-shield]][github-action-url]
[![Coverage Status][coveralls-shield]][coveralls-url]
[![Go Report Card][reportcard-shield]][reportcard-url]
[![Issues][issues-shield]][issues-url]
[![Issues][stars-shield]][stars-url]




<p align="center">
  <h2 align="center">Message Interface</h2>

  <p align="center">
    An interface to add domain events and messages to your Go projects!
    <br />
    <br />
    <a href="https://github.com/HTechHQ/message#examples">View Examples</a>
    ·
    <a href="https://github.com/HTechHQ/message/issues">Report Bug</a>
    ·
    <a href="https://github.com/HTechHQ/message/issues">Request Feature</a>
  </p>
</p>




## About the Project
### Why?
As a project matures, requirements and the technical 
environment evolve, it is painful to realise that deeply embedded
dependencies like loggers or message systems have to be replaced.
This is especially true for Go, that is not offering standard
abstractions for these use cases.
Thus, it is a helpful practise to hide external libraries behind an
application specific interface.

This is such an (opinionated) interface intended to be **copied**
into your project.

### Principles & Design Goals
The goal is a message interface, that can be implemented by multiple libraries
as time progresses.\
This repository explores what a good interface looks like for Go and
provides an in memory reference implementation.

1. **Simplicity.**
   The most important part is the simple interface. It provides just enough
   to send and receive messages in an application.
1. **Change resilient.**
   Don't just import this project. 
   Copy the interface into your app and use any of the 
   available implementations, so your use cases and domain don't depend on any
   library in the future. Not even on this one.
1. **Developer convenience.**
   With an emphasis on development speed and readable code 
   this interface improves upon the pattern often seen with other libraries:
   marshalling and unmarshalling the message to and from `[]byte`.
   This clutters your code with serialisation logic 
   and can be hidden behind an interface.




## [The Interface](pubsub.go)
```go
type PubSub interface {
    Publish(ctx context.Context, eom EventOrMessage) error
    Subscribe(eom EventOrMessage, h HandlerFunc) (*Subscriber, error)
    Shutdown(ctx context.Context)
}
```




## Examples
```shell
$ go get -u github.com/HTechHQ/message
```


[Publish-Subscribe with messages](examples/pubsub-message/main.go)
```go
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

```


[Publish-Subscribe with domain events](examples/pubsub-events/main.go)
```go
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

// newUserRegistered is the actual event that's fired after a new user got registered.
type newUserRegistered struct {
    Username  string
    Email     string
    CreatedAt time.Time
    Settings  []string
}
```

See [all examples](examples/)




## Roadmap
Depending on your needs a persistent implementation, relying on something like
NATS or Redis, might be good to support your desired semantics 
like at-least-once-delivery. PR welcome.

Also working on metrics and a queue expansion.
\
\
See the [open issues](https://github.com/HTechHQ/message/issues) for a list of proposed features (and known issues).




<!-- MARKDOWN LINKS & IMAGES -->
[issues-shield]: https://img.shields.io/github/issues/HTechHQ/message?style=flat-square&logo=appveyor
[issues-url]: https://github.com/HTechHQ/message/issues
[stars-shield]: https://img.shields.io/github/stars/HTechHQ/message?style=flat-square&logo=appveyor
[stars-url]: https://github.com/HTechHQ/message/stargazers
[reportcard-shield]: https://goreportcard.com/badge/github.com/HTechHQ/message
[reportcard-url]: https://goreportcard.com/report/github.com/HTechHQ/message
[coveralls-shield]: https://coveralls.io/repos/github/HTechHQ/message/badge.svg?branch=master
[coveralls-url]: https://coveralls.io/github/HTechHQ/message?branch=master
[github-action-shield]: https://github.com/HTechHQ/message/actions/workflows/test.yml/badge.svg
[github-action-url]: https://github.com/HTechHQ/message/actions