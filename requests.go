package httpclient

import (
	"context"
	"io"
	"net/http"
)

// Do executes specified HTTP method with provided body (if not nil) and
// then calls the response handler with the result. This method can be used to
// create custom HTTP requests (e.g. PATCH)
func (c *client) Do(
	ctx context.Context,
	onReponse ResponseHandler,
	method, url string,
	body io.Reader,
	opts ...RequestOption,
) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	// copy headers from client
	for k, v := range c.headers {
		for _, dv := range v {
			req.Header.Add(k, dv)
		}
	}
	// apply any request options that may have been passed
	for _, opt := range opts {
		if err := opt(req); err != nil {
			return err
		}
	}
	// make the request and return the response
	res, err := c.client.Do(req)
	if res != nil {
		if res.Body != nil {
			defer res.Body.Close() // idempotent
		}
	}
	return onReponse(ctx, res, err)
}

// Get executes a get request and calls the response handler with the result
func (c *client) Get(
	ctx context.Context,
	onResponse ResponseHandler,
	url string,
	opts ...RequestOption,
) error {
	return c.Do(ctx, onResponse, "GET", url, nil, opts...)
}

// Post executes a POST request with provided body and then calls the
// response handler with the result
func (c *client) Post(
	ctx context.Context,
	onResponse ResponseHandler,
	url string,
	body io.Reader,
	opts ...RequestOption,
) error {
	return c.Do(ctx, onResponse, "POST", url, body, opts...)
}
