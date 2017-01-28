package httpclient

import (
	"context"
	"io"
	"net/http"
)

// Do executes specified HTTP method with provided requestBody (if not nil)
// and then calls the response handler with the result.
// This method can be used to create custom HTTP requests (e.g. PATCH).
func (c *Client) Do(
	ctx context.Context,
	rh ResponseHandler,
	method, url string,
	requestBody io.Reader,
	opts ...RequestOption) error {
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	//copy headers from client
	for k, v := range c.headers {
		for _, dv := range v {
			req.Header.Add(k, dv)
		}
	}
	//apply any request options that may have been passed
	for _, optFunc := range opts {
		if err := optFunc(req); err != nil {
			return err
		}
	}
	resp, err := c.client.Do(req)
	if resp != nil {
		if resp.Body != nil {
			defer resp.Body.Close() //idempotent
		}
	}
	return rh(ctx, resp, err)
}
