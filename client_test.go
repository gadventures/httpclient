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
	h2c, err := New(
		DialTimeout(3*time.Second),
		Headers(headers),
		IdleConnTimeout(30*time.Second),
		KeepAliveTimeout(60*time.Second),
		Logger(ioutil.Discard),
		MaxIdleConns(4),
		MaxIdleConnsPerHost(2),
		RedirectPolicy(defaultRedirectPolicy),
		ResponseHeaderTimeout(20*time.Second),
		TLSHandshakeTimeout(10*time.Second),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}

	tests := []struct {
		opt    func(c *client) error
		errstr string
	}{
		{badConfigOption(), "badd"},
		{MaxIdleConns(-1), ErrInvalidValue.Error()},
		{MaxIdleConnsPerHost(-2), ErrInvalidValue.Error()},
	}
	for _, test := range tests {
		_, err := New(test.opt)
		if err.Error() != test.errstr {
			t.Errorf("expected %s but got %s", test.errstr, err.Error())
		}
	}

	if h2c.Client() == nil {
		t.Errorf("Expected non nil *http.Client")
	}
}

func badConfigOption() func(c *client) error {
	return func(c *client) error {
		return errors.New("badd")
	}
}
