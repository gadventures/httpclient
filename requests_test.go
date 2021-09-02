package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	headers := make(http.Header)
	headers.Add("X-Test", "TestGet")
	c, err := New(
		DialTimeout(4*time.Second),
		Headers(headers),
		IdleConnTimeout(10*time.Second),
		Logger(os.Stderr),
		MaxIdleConns(4),
		RedirectPolicy(defaultRedirectPolicy),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}
	ctx := context.Background()

	rh := func(ctx context.Context, resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stderr, resp.Body)
		return err
	}

	extra := make(http.Header)
	extra.Add("X-Extra", "true")
	delheaders := make(http.Header)
	delheaders.Add("X-Test", "")
	err = c.Get(ctx, rh, "https://httpbin.org/get", AddHeaders(extra), DelHeaders(delheaders))
	if err != nil {
		t.Errorf("trouble when making GET request: %v", err)
	}

	// test something fake
	err = c.Get(ctx, rh, "https://bing.bang.foo.bar.moo.moo/get")
	if err == nil {
		t.Error("Expected an error but got nothing")
	}

	// erroring request option
	err = c.Get(ctx, rh, "https://httpbin.org/get", func(req *http.Request) error {
		return errors.New("baad")
	})
	if err == nil {
		t.Error("Expected an error but got nothing")
	}

	// test codes and errors
	var wg sync.WaitGroup
	tests := []struct {
		url string
		f   ResponseHandler
	}{
		{"https://httpbin.org/status/404", codeTest(&wg, t, 404)},
		{"https://httpbin.org/status/401", codeTest(&wg, t, 401)},
		{"https://httpbin.org/status/500", codeTest(&wg, t, 500)},
		{"https://httpbin.org/status/403", codeTest(&wg, t, 403)},
		{"https://httpbin.org/status/400", codeTest(&wg, t, 400)},
		{"https://httpbin.org/status/409", codeTest(&wg, t, 409)},
	}
	for _, test := range tests {
		wg.Add(1)
		test := test
		go func() {
			c.Get(ctx, test.f, test.url)
		}()
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	c.Close()
}

func codeTest(wg *sync.WaitGroup, t *testing.T, code int) ResponseHandler {
	return func(ctx context.Context, resp *http.Response, err error) error {
		if err != nil {
			t.Errorf("GET should have suceeded: %v", err)
		}
		defer wg.Done()
		if resp.StatusCode != code {
			t.Errorf("unexpected status in response: %v", resp.Status)
		}
		return nil
	}
}

func TestGetH2(t *testing.T) {
	c, err := New(
		DialTimeout(3*time.Second),
		IdleConnTimeout(10*time.Second),
		Logger(os.Stderr),
		LogPrefix("h2client-test: "),
		MaxIdleConns(4),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}
	ctx := context.Background()

	rh := func(ctx context.Context, resp *http.Response, err error) error {
		if !strings.Contains(resp.Proto, "HTTP/2") {
			return fmt.Errorf("Unexpected protocol %s wanted HTTP/2.0", resp.Proto)
		}
		if err != nil {
			return err
		}
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		sleep := time.Duration(i) * time.Millisecond * 100
		go func() {
			time.Sleep(sleep)
			err := c.Get(ctx, rh, "https://http2.akamai.com/demo")
			if err != nil {
				t.Errorf("trouble when making GET request: %v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	c.Close()
}

func TestGetDisableH2(t *testing.T) {
	c, err := New(
		DialTimeout(3*time.Second),
		DisableHTTP2(),
		IdleConnTimeout(10*time.Second),
		Logger(os.Stderr),
		LogPrefix("h2client-test: "),
		MaxIdleConns(4),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}
	ctx := context.Background()

	rh := func(ctx context.Context, resp *http.Response, err error) error {
		if strings.Contains(resp.Proto, "HTTP/2") {
			return fmt.Errorf("Unexpected protocol %s wanted HTTP/1.1", resp.Proto)
		}
		if err != nil {
			return err
		}
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		sleep := time.Duration(i) * time.Millisecond * 100
		go func() {
			time.Sleep(sleep)
			err := c.Get(ctx, rh, "https://http2.akamai.com/demo")
			if err != nil {
				t.Errorf("trouble when making GET request: %v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	c.Close()
}

func TestPost(t *testing.T) {
	headers := make(http.Header)
	headers.Add("X-Test", "TestPost")
	headers.Add("Content-Type", "application/json")
	c, err := New(
		DialTimeout(4*time.Second),
		Headers(headers),
		IdleConnTimeout(10*time.Second),
		Logger(os.Stderr),
		MaxIdleConns(4),
		RedirectPolicy(defaultRedirectPolicy),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}
	defer c.Close()
	ctx := context.Background()

	rh := func(ctx context.Context, resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stderr, resp.Body)
		return err
	}

	buf := bytes.NewBufferString(`{"foo": "bar"}`)
	extra := make(http.Header)
	extra.Add("X-Extra", "true")
	err = c.Post(ctx, rh, "https://httpbin.org/post", buf, SetHeaders(extra))
	if err != nil {
		t.Errorf("trouble when making POST request: %v", err)
	}
}
