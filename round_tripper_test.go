package httpclient

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestCustomRoundTripper(t *testing.T) {
	// headers
	headers := make(http.Header)
	headers.Add("X-Test", "TestClient")

	// client & test
	c, err := New(
		Headers(headers),
		WithRoundTripper(RoundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errNone
		})),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}
	defer c.Close()

	err = c.Get(context.Background(), NoopResponseHandler, "https://www.gadventures.com")
	if !errors.Is(err, errNone) {
		t.Errorf("trouble with custom roundtripper unexpected -> %v", err)
	}
}
