package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestPost(t *testing.T) {
	headers := make(http.Header)
	headers.Add("X-Test", "TestPost")
	headers.Add("Content-Type", "application/json")
	c, err := New(
		Headers(headers),
		DialTimeout(4*time.Second),
		IdleConnTimeout(10*time.Second),
		MaxIdleConns(4),
		Logger(os.Stderr),
		RedirectPolicy(defaultRedirectPolicy),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}
	defer c.Close()
	ctx := context.Background()

	rh := func(ctx context.Context, resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stderr, resp.Body)
		return err
	}

	buf := bytes.NewBufferString(`{"foo": "bar"}`)
	extra := make(http.Header)
	extra.Add("X-Extra", "true")
	err = c.Post(ctx, rh, "https://httpbin.org/post", buf, ExtraHeaders(extra))
	if err != nil {
		t.Errorf("trouble when making POST request: %v", err)
	}

}
