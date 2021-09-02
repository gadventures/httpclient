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
	// default headers used to init the Client
	headers := make(http.Header)
	headers.Add("X-Test", "TestClient")

	// test case struct
	type testCase struct {
		name     string
		extra    map[string][]string
		expected map[string][]string
	}

	// table tests
	for _, tc := range []testCase{
		{
			name: "default case",
			expected: map[string][]string{
				"X-Test": {"TestClient"},
			},
		},
		{
			name: "overriding X-Test header value",
			extra: map[string][]string{
				"X-Test": {"Override"},
			},
			expected: map[string][]string{
				"X-Test": {"Override"},
			},
		},
		{
			name: "adding X-Extra with single value",
			extra: map[string][]string{
				"X-Extra": {"true"},
			},
			expected: map[string][]string{
				"X-Extra": {"true"},
				"X-Test":  {"TestClient"},
			},
		},
		{
			name: "adding X-Extra with multiple values",
			extra: map[string][]string{
				"X-Extra": {"true", "multi"},
			},
			expected: map[string][]string{
				"X-Extra": {"true", "multi"},
				"X-Test":  {"TestClient"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := New(Headers(headers))
			if err != nil {
				t.Errorf("trouble when creating the client: %v", err)
			}
			defer c.Close()

			// startup a server with handler to check that received request headers
			// are correct
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// ensure headers are _received_ as expected
				for k, vs := range tc.expected {
					exp := strings.Join(vs, ", ")
					act := strings.Join(r.Header.Values(k), ", ")
					if act != exp {
						t.Errorf("expected \"%s: %s\" Header value, got \"%s: %s\"", k, exp, k, act)
					}
				}
			}))
			defer s.Close()

			// create our extra headers
			extra := make(http.Header)
			for k, vs := range tc.extra {
				for _, v := range vs {
					extra.Add(k, v)
				}
			}
			// make the request
			err = c.Get(context.Background(), NoopResponseHandler, s.URL, SetHeaders(extra))
			if err != nil {
				t.Errorf("unexpected error with SetHeaders: %v", err)
			}
		})
	}
}
