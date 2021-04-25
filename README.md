
[![Issues][issues-shield]][issues-url]
[![Issues][stars-shield]][stars-url]




<p align="center">
  <h2 align="center">Message Interface</h2>

  <p align="center">
    An interface to add messages to your Go projects!
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
   This clutters the use cases with serialisation logic 
   and can be hidden behind an interface.




## The Interface
```go

```




## Examples




## Roadmap
Depending on your needs a persistent implementation, relying on something like
NATS or Redis, might be good to support your desired semantics 
like at-least-once-delivery. PR welcome.

Also working on an expansion to a queue.
\
\
See the [open issues](https://github.com/HTechHQ/message/issues) for a list of proposed features (and known issues).




<!-- MARKDOWN LINKS & IMAGES -->
[issues-shield]: https://img.shields.io/github/issues/HTechHQ/message?style=flat-square&logo=appveyor
[issues-url]: https://github.com/HTechHQ/message/issues
[stars-shield]: https://img.shields.io/github/stars/HTechHQ/message?style=flat-square&logo=appveyor
[stars-url]: https://github.com/HTechHQ/message/stargazers
