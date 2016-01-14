# Napping: HTTP for Gophers

Package `napping` is a [Go][] client library for interacting with
[RESTful APIs][].  Napping was inspired  by Python's excellent [Requests][]
library.


## Status

[![Drone Build Status](https://drone.io/github.com/jmcvetta/napping/status.png)](https://drone.io/github.com/jmcvetta/napping/latest)
[![Travis Build Status](https://travis-ci.org/jmcvetta/napping.png)](https://travis-ci.org/jmcvetta/napping)
[![Coverage Status](https://coveralls.io/repos/jmcvetta/restclient/badge.png)](https://coveralls.io/r/jmcvetta/napping)

Used by, and developed in conjunction with, [Neoism][].


## Installation 

### Requirements

Napping requires Go 1.2 or later.


### Development

```
go get github.com/jmcvetta/napping
```

### Stable

Napping is versioned using [`gopkg.in`](http://gopkg.in).  

Current release is `v3`.

```
go get gopkg.in/jmcvetta/napping.v3
```


## Documentation

See [![GoDoc](http://godoc.org/github.com/jmcvetta/napping?status.png)](http://godoc.org/github.com/jmcvetta/napping)
for automatically generated API documentation.

Check out [github_auth_token][auth-token] for a working example
showing how to retrieve an auth token from the Github API.


## Production Note

If you decide to use Napping in a production system please let me know.  All
API changes will be made via Pull Request, so it's highly recommended you Watch
the repo Issues.  The API is fairly stable but there may be additions and small 
changes from time to time.


## Contributing

Contributions in the form of Pull Requests or Issues are gladly accepted.
Before submitting a Pull Request, please ensure your code passes all tests, and
that your changes do not decrease test coverage.  I.e. if you add new features,
also add corresponding new tests.

When submitting an Issue, if possible please include a failing test case that 
demonstrates the problem.


## License

This is Free Software, released under the terms of the [GPL v3][].

Please feel free to make a donation to the author, to support the development
of this and other Free Software packages.

[![Donate with PayPal](https://img.shields.io/badge/donate-paypal-blue.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=YEXKK27UL48F2)


[Go]:           http://golang.org
[RESTful APIs]: http://en.wikipedia.org/wiki/Representational_state_transfer#RESTful_web_APIs
[Requests]:     http://python-requests.org
[GPL v3]:       http://www.gnu.org/copyleft/gpl.html
[auth-token]:   https://github.com/jmcvetta/napping/blob/master/examples/github_auth_token/github_auth_token.go
[Neoism]:       https://github.com/jmcvetta/neoism
