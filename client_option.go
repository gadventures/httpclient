package httpclient

import (
	"io"
	"net/http"
	"time"
)

// Option is our functional options type
// see: https://sagikazarmark.hu/blog/functional-options-on-steroids/
type Option func(*client) error

// DialTimeout is configuration option to pass to client it changes how long
// the client will wait to establish the TCP connection
func DialTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.dialTimeout = t
		return nil
	}
}

// DisableHTTP2 is configuration option to pass to client. It disables the
// Transport support for HTTP/2, thereby forcing HTTP/1.1.
func DisableHTTP2() Option {
	return func(c *client) error {
		c.disableHTTP2 = true
		return nil
	}
}

// DisableKeepAlive is configuration option to pass to client
// it disables KeepAlive for the tcp connection
func DisableKeepAlive() Option {
	return func(c *client) error {
		c.disableKeepAlive = true
		return nil
	}
}

// Headers is configuration option to pass headers to Client. It makes GET
// requests using the provided headers. Use this for headers that are to be
// shared among all Requests. For Request specific headers use the
// SetHeaders RequestOption.
func Headers(headers http.Header) Option {
	return func(c *client) error {
		for k, v := range headers {
			c.headers[k] = v
		}
		return nil
	}
}

// IdleConnTimeout is a configuration option to pass to client. It sets how
// long idle connections waiting to be used again live in a pool. Once the
// timeout is reached connections are closed and removed from pool.
func IdleConnTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.idleConnTimeout = t
		return nil
	}
}

// KeepAliveTimeout is configuration option to pass to client. It changes the
// timeout for how long the TCP Connection stays open, after which period will
// TCP keep alive happen on the TCP connection. It is safe to leave the default
// value, but can be tweaked should one notice connections being reset by peers
// sooner than expected.
//
// NOTE: The OS may override this value
func KeepAliveTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.keepAliveTimeout = t
		return nil
	}
}

// Logger is configuration option to pass to client. It changes where the debug
// logs are written (ioutil.Discard by default).
func Logger(w io.Writer) Option {
	return func(c *client) error {
		c.logWriter = w
		return nil
	}
}

// LogPrefix is configuration option to pass to client. It will change the
// prefix used in Client's log output. This can be useful when one is using
// several httpclients
func LogPrefix(p string) Option {
	return func(c *client) error {
		c.logPrefix = p
		return nil
	}
}

// MaxIdleConns is configuration option to pass to client. It changes the
// maximum idle connections that the client will keep in a pool.
func MaxIdleConns(n int) Option {
	return func(c *client) error {
		if n < 0 {
			return ErrInvalidValue
		}
		c.maxIdleConns = n
		return nil
	}
}

// MaxIdleConnsPerHost is configuration option to pass to client. It changes
// the maximum idle connections that the client will keep in a pool for a
// single host. If not specified this will be the same as MaxIdleConns.
func MaxIdleConnsPerHost(n int) Option {
	return func(c *client) error {
		if n < 0 {
			return ErrInvalidValue
		}
		c.maxIdleConnsPerHost = n
		return nil
	}
}

// RedirectPolicy is configuration option to pass to client. It changes what
// the client does on redirects. The default behaviour is to copy the original
// request headers and try again up to 10 times.
func RedirectPolicy(redirectFunc func(req *http.Request, via []*http.Request) error) Option {
	return func(c *client) error {
		c.redirectFunc = redirectFunc
		return nil
	}
}

// TLSHandshakeTimeout is a configuration option to pass to client. It limits
// the time spent performing the TLS handshake.
func TLSHandshakeTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.tlsHandshakeTimeout = t
		return nil
	}
}

// ResponseHeaderTimeout is a configuration option to pass to client. It limits
// the time spent reading the headers of the response.
func ResponseHeaderTimeout(t time.Duration) Option {
	return func(c *client) error {
		c.responseHeaderTimeout = t
		return nil
	}
}

// RoundTripperFunc is like http.HandlerFunc, but for RoundTripper interface.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip satisfies the http.RoundTripper interface.
func (r RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

// WithRoundTripper is configuration option to pass to client. This will change
// the http.RoundTripper that the client will use.
//
// NOTE: the usage of this renders the use of httpclient pointless, because if
//       you are managing your own transports, you might as well use net/http
//       directly. This option is useful for unittests, such as when using
//       httpmock to verify expected traffic when testing some code that uses
//       httpclient.
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
