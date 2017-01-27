package httpclient

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	headers := make(http.Header)
	headers.Add("X-Test", "TestClient")
	_, err := New(
		Headers(headers),
		KeepAliveTimeout(60*time.Second),
		DialTimeout(3*time.Second),
		MaxIdleConns(2),
		Logger(ioutil.Discard),
		RedirectPolicy(defaultRedirectPolicy),
		IdleConnTimeout(30*time.Second),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}

	_, err = New(func(c *Client) error {
		return errors.New("badd")
	})
	if err == nil {
		t.Error("expected error on bad option")
	}

}
