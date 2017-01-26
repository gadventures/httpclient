package httpclient

import (
	"context"
	"net/http"
)

//Get executes a get request and calls the response handler with the result
func (c *Client) Get(ctx context.Context, rh ResponseHandler, url string, opts ...RequestOption) error {
	req, err := http.NewRequest("GET", url, nil)
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
