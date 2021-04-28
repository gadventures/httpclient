package httpclient

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestCustomRoundTripper(t *testing.T) {
	headers := make(http.Header)
	headers.Add("X-Test", "TestClient")
	c, err := New(
		Headers(headers),
		WithRoundTripper(&TestRoundTripper{}),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}
	defer c.Close()

	rh := func(ctx context.Context, resp *http.Response, err error) error {
		return err
	}
	err = c.Get(context.Background(), rh, "http://www.gadventures.com")
	if err == nil || !strings.Contains(err.Error(), "was here fool") {
		t.Errorf("trouble with custom roundtripper unexpected -> %v", err)
	}
}

type TestRoundTripper struct{}

func (t *TestRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("was here fool")
}
