package httpclient

import (
	"io"
	"net/http"
	"time"
)

// Option is our functional options type
// see: https://sagikazarmark.hu/blog/functional-options-on-steroids/
type Option func(*client) error

// DialTimeout is configuration option to pass to Client it changes how long
// the client will wait to establish the TCP connection
func DialTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.dialTimeout = t
		return nil
	}
}

// DisableHTTP2 is configuration option to pass to Client
// it disables the transparent support for HTTP/2
// thereby forcing HTTP/1.1
func DisableHTTP2() Option {
	return func(c *client) error {
		c.disableHTTP2 = true
		return nil
	}
}

// DisableKeepAlive is configuration option to pass to Client
// it disables KeepAlive for the tcp connection
func DisableKeepAlive() Option {
	return func(c *client) error {
		c.disableKeepAlive = true
		return nil
	}
}

// Headers is configuration option to pass headers to Client
// it makes GET requests use the provided headers
// use this for headers shared among all requests
// for request specific headers use the ExtraHeaders RequestOption
func Headers(headers http.Header) Option {
	return func(c *client) error {
		for k, v := range headers {
			c.headers[k] = v
		}
		return nil
	}
}

// IdleConnTimeout is a configuration option to pass to Client
// it sets how long will idle connections live in a pool
// waiting to be used again.
// Once the timeout is reached connections are closed and removed from pool.
func IdleConnTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.idleConnTimeout = t
		return nil
	}
}

// KeepAliveTimeout is configuration option to pass to Client
// it changes the timeout for how long the tcp connection stays open
// After which period will tcp keep alive happen on the TCP connection
// safe to leave the default but can be tweaked should one notice
// connections being reset by peers sooner than expected
// Note: OS may override this value
func KeepAliveTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.keepAliveTimeout = t
		return nil
	}
}

// Logger is configuration option to pass to Client
// it changes where the debug info is written
// (ioutil.Discard by default)
func Logger(w io.Writer) Option {
	return func(c *client) error {
		c.logWriter = w
		return nil
	}
}

// LogPrefix is configuration option to pass to Client
// to change the prefix used in Clients log output.
// This can be useful when one is using several httpclients.
func LogPrefix(p string) Option {
	return func(c *client) error {
		c.logPrefix = p
		return nil
	}
}

// MaxIdleConns is configuration option to pass to Client
// it changes the maximum idle connections that the client will keep
// in a pool
func MaxIdleConns(n int) Option {
	return func(c *client) error {
		if n < 0 {
			return ErrInvalidValue
		}
		c.maxIdleConns = n
		return nil
	}
}

// MaxIdleConnsPerHost is configuration option to pass to Client
// it changes the maximum idle connections that the client will keep in a pool
// for a single host.
// If not specified this will be the same as MaxIdleConns.
func MaxIdleConnsPerHost(n int) Option {
	return func(c *client) error {
		if n < 0 {
			return ErrInvalidValue
		}
		c.maxIdleConnsPerHost = n
		return nil
	}
}

// RedirectPolicy is configuration option to pass to Client
// it changes what client does on redirects
// the default behaviour is to copy the headers from original
// request and try again up to 10 times
func RedirectPolicy(redirectFunc func(req *http.Request, via []*http.Request) error) Option {
	return func(c *client) error {
		c.redirectFunc = redirectFunc
		return nil
	}
}

// TLSHandshakeTimeout is a configuration option to pass to Client
// it limits the time spent performing the TLS handshake.
func TLSHandshakeTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.tlsHandshakeTimeout = t
		return nil
	}
}

// ResponseHeaderTimeout is a configuration option to pass to Client
// it limits the time spent reading the headers of the response.
func ResponseHeaderTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.responseHeaderTimeout = t
		return nil
	}
}

// WithRoundTripper is configuration option to pass to Client
// this will change the http.RoundTripper that the client will use
// note use of this renders the use of httpclient pointless
// since if you are managing your own transports
// you might as well use net/http directly
// however this option is useful for unittests
// e.g. when using httpmock to verify expected traffic
// when testing some code that uses httpclient
func WithRoundTripper(rt http.RoundTripper) Option {
	return func(c *client) error {
		c.customRoundTripper = rt
		return nil
	}
}

// set the Options provided to the New method
func (c *client) setOptions(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}
