### HTTPClient

[![GoDoc](https://godoc.org/github.com/gadventures/httpclient?status.svg)](https://godoc.org/github.com/gadventures/httpclient)   

Thin layer around standard HTTP go library with extensible configuration

A lot of the http code around making requests to our REST APIs can be extracted and shared among various projects.

This includes
* Safe for concurrency
* Transparent support for HTTP/2
* Header sharing between requests
* Custom redirect policies
* Calling GET requests with context objects
* Ability to set and share various timeouts without diving deep into net/http
* Having better understanding regarding idle connection pools

All of the above is supported by standard library, however this thin layer allows to do so without having to manage own
http.Client and http.Transport objects.


```golang

import htc "github.com/gadventures/httpclient"

client, err := htc.New(
   htc.KeepAliveTimeout(30*time.Second),
   htc.DialTimeout(2*time.Second),
   htc.MaxIdleConns(4),
   htc.Logger(os.Stderr),
)

responseHandler := func(ctx context.Context, resp *http.Response, err error) error {
   if err != nil {
      return err
   }
   _, err = io.Copy(os.Stdout, resp.Body)
   return err
}

err = client.Get(ctx, responseHandler, "https://somewhere.foo.bar.com/page.html")

```


Alpha quality code - use at your own risk.
