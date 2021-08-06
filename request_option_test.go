package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var errNone = errors.New("no error")

func TestSetHeaders(t *testing.T) {
	// default headers
	headers := make(http.Header)
	headers.Add("X-Test", "TestClient")

	// test case struct
	type tCase struct {
		exp map[string][]string
	}

	// table tests
	for _, tc := range []tCase{
		{map[string][]string{
			"X-Test": {"Override"},
		}},
		{map[string][]string{
			"X-Extra": {"true"},
		}},
		{map[string][]string{
			"X-Extra": {"true", "multi"},
		}},
	} {
		c, err := New(Headers(headers))
		if err != nil {
			t.Errorf("trouble when creating the client: %v", err)
		}

		// startup a server with handler to check that received request headers
		// are correct
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// add X-Test if we're not already checking for it
			if _, ok := tc.exp["X-Test"]; !ok {
				tc.exp["X-Test"] = headers.Values("X-Test")
			}
			// ensure headers are _received_ as expected
			for k, vs := range tc.exp {
				exp := strings.Join(vs, ", ")
				act := strings.Join(r.Header.Values(k), ", ")
				if act != exp {
					t.Errorf("missing \"%s: %s\" header, got %s", k, exp, act)
				}
			}
		}))

		// extra headers
		extra := make(http.Header)
		for k, vs := range tc.exp {
			for _, v := range vs {
				extra.Add(k, v)
			}
		}
		// make the request
		err = c.Get(context.Background(), NoopResponseHandler, s.URL, SetHeaders(extra))
		if err != nil {
			t.Errorf("trouble with set headers roundtripper unexpected -> %v", err)
		}
		// close client & server
		c.Close()
		s.Close()
	}
}
