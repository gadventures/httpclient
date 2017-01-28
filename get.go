package httpclient

import "context"

//Get executes a get request and calls the response handler with the result
func (c *Client) Get(ctx context.Context, rh ResponseHandler, url string, opts ...RequestOption) error {
	return c.Do(ctx, rh, "GET", url, nil, opts...)
}
