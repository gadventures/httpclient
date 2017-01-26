package httpclient

import (
	"fmt"
	"net/http"
)

func defaultRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) > 10 {
		return fmt.Errorf("too many redirects")
	}

	//copy headers to current redirect
	if len(via) > 0 {
		for k, v := range via[0].Header {
			req.Header[k] = v
		}
	}
	return nil
}
