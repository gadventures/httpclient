package httpclient

import (
	"net/http"
	"net/url"
)

// RequestOption allows additional modifications to request object before
// http.Do is called
type RequestOption func(*http.Request) error

// AddHeaders allows for additional headers to be added when making a request
func AddHeaders(headers http.Header) RequestOption {
	return func(req *http.Request) error {
		for k, v := range headers {
			for _, dv := range v {
				req.Header.Add(k, dv)
			}
		}
		return nil
	}
}

// DelHeaders allows for certain headers to be removed when making requests
func DelHeaders(headers http.Header) RequestOption {
	return func(req *http.Request) error {
		for k := range headers {
			req.Header.Del(k)
		}
		return nil
	}
}

// SetHeaders allows for certain headers to be replaced when making a request
func SetHeaders(headers http.Header) RequestOption {
	return func(req *http.Request) error {
		for k, vs := range headers {
			req.Header.Del(k)
			for i := range vs {
				req.Header.Add(k, vs[i])
			}
		}
		return nil
	}
}

// AddQueryParams is an additive option that will not replace existing query
// parameters
func AddQueryParams(values url.Values) RequestOption {
	return func(req *http.Request) error {
		params := req.URL.Query()
		for k, vals := range values {
			for _, v := range vals {
				params.Add(k, v)
			}
		}
		req.URL.RawQuery = params.Encode()
		return nil
	}
}

// WithQueryParams will override any existing query parameters bound to the
// request
func WithQueryParams(values url.Values) RequestOption {
	return func(req *http.Request) error {
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
