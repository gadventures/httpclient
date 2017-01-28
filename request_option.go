package httpclient

import "net/http"

// RequestOption allows additional modifications
// to request object before http.Do is called
type RequestOption func(*http.Request) error

// AddHeaders allows for additional headers
// to be added when making a request
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

// DelHeaders allows for certain headers
// to be removed when making requests
func DelHeaders(headers http.Header) RequestOption {
	return func(req *http.Request) error {
		for k := range headers {
			req.Header.Del(k)
		}
		return nil
	}
}

// SetHeaders allows for certain headers
// to be replaced when making a request
func SetHeaders(headers http.Header) RequestOption {
	return func(req *http.Request) error {
		for k, v := range headers {
			var val []string
			copy(val, v)
			req.Header[k] = val
		}
		return nil
	}
}
