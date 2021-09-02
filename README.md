HTTPClient [![Go Report Card](https://goreportcard.com/badge/github.com/gadventures/httpclient)](https://goreportcard.com/report/github.com/gadventures/httpclient) [![Go Reference](https://pkg.go.dev/badge/github.com/gadventures/httpclient.svg)](https://pkg.go.dev/github.com/gadventures/httpclient)
==========

Thin layer around Go's `net/http` library with extensible configuration options.

It extracts the http code making requests to our REST APIs to be shared among various projects.

Features include:

* Safety for concurrent usage
* Transparent support for HTTP/2
* Sharing Headers between requests
* Custom redirect policies
* Calling GET requests with context objects
* Ability to set and share various timeouts without diving deep into `net/http` internals
* Having a better understanding regarding idle connection pools

All of the above is supported by the standard library, however this thin layer allows us to do so without having to manage our own `http.Client` and `http.Transport` objects.

For more information on HTTP Timeouts [read this](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/).

Usage
-----

```go
package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	h2 "github.com/gadventures/httpclient"
)

func main() {
	client, err := h2.New(
		h2.DialTimeout(3*time.Second),
		h2.IdleConnTimeout(30*time.Second),
		h2.Logger(os.Stderr),
		h2.MaxIdleConns(4),
	)

	responseHandler := func(ctx context.Context, resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	}

	err = client.Get(ctx, responseHandler, "https://somewhere.foo.bar.com/page.html")
}
```

**WARNING:** alpha quality code; use at your own risk.
