package httpclient

import (
	"context"
	"io"
)

// Post executes a POST request with provided requestBody
// and then calls the response handler with the result.
func (c *Client) Post(
	ctx context.Context,
	rh ResponseHandler,
	url string,
	requestBody io.Reader,
	opts ...RequestOption) error {
	return c.Do(ctx, rh, "POST", url, requestBody, opts...)
}
