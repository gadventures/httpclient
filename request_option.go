package httpclient

import "net/http"

// RequestOption allows additional modifications
// to request object before http.Do is called
type RequestOption func(*http.Request) error

// ExtraHeaders allows for additional headers
// to be passed to e.g. GET requests
func ExtraHeaders(headers http.Header) RequestOption {
	return func(req *http.Request) error {
		for k, v := range headers {
			for _, dv := range v {
				req.Header.Add(k, dv)
			}
		}
		return nil
	}
}
