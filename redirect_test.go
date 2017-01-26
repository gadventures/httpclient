package httpclient

import (
	"net/http"
	"testing"
)

func TestRedirect(t *testing.T) {

	vreq, err := http.NewRequest("GET", "https://www.gadventures.com", nil)
	vreq.Header.Add("foo", "bar")
	vreq.Header.Add("moo", "baah")
	via := []*http.Request{vreq}

	req, err := http.NewRequest("GET", "https://www.gadventures.com", nil)
	err = defaultRedirectPolicy(req, via)
	if err != nil {
		t.Errorf("redirect failed: %v", err)
	}
	if req.Header.Get("foo") != "bar" || req.Header.Get("moo") != "baah" {
		t.Error("expected headers missing from redirect")
	}
	for i := 0; i < 10; i++ {
		via = append(via, vreq)
	}
	err = defaultRedirectPolicy(req, via)
	if err == nil {
		t.Error("expecteded error was nil")
	} else {
		if err.Error() != "too many redirects" {
			t.Errorf(`Expected "too many redirects" but got "%s"`, err.Error())
		}
	}
}
