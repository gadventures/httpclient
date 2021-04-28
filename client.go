package httpclient

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	// DefaultDialTimeout for TCP dial
	DefaultDialTimeout = 10 * time.Second

	// DefaultTLSHandshakeTimeout for TLS handshake
	DefaultTLSHandshakeTimeout = 10 * time.Second

	// DefaultResponseHeaderTimeout  for waiting to read a response header
	DefaultResponseHeaderTimeout = 30 * time.Second

	// DefaultMaxIdleConns to keep in pool
	DefaultMaxIdleConns = 15

	// DefaultKeepAliveTimeout - when socket keep alive check will be performed
	DefaultKeepAliveTimeout = 90 * time.Second

	// log-prefix
	defaultLogPrefix = "[httpclient]: "
)

// ErrInvalidValue signifies that an invaild value was given to configartion option
var ErrInvalidValue = errors.New("invalid value for option")

type Client interface {
	// Client returns the underlying *http.Client
	Client() *http.Client

	// Close all Idle connections
	Close()

	// HTTP Request Methods //

	// Do is the generic HTTP request method. The two string parameters in
	// order are the HTTP Method, and the URL to request respectively
	Do(context.Context, ResponseHandler, string, string, io.Reader, ...RequestOption) error

	// Get some URL
	Get(context.Context, ResponseHandler, string, ...RequestOption) error

	// POST to some URL
	Post(context.Context, ResponseHandler, string, io.Reader, ...RequestOption) error
}

// ensure interface implementation
var _ Client = &client{}

// Client is our http client that can be used to make http requests efficiently
// for now we intend to support Get
// it is used rather than net/http package directly due to ability to configure
// timeouts and watch the resource use
// safe (and intended) to use from several go routines
type client struct {
	client                *http.Client
	currentConnID         int64
	customRoundTripper    http.RoundTripper
	dialTimeout           time.Duration
	disableHTTP2          bool
	disableKeepAlive      bool
	headers               http.Header
	idleConnTimeout       time.Duration
	keepAliveTimeout      time.Duration
	log                   *log.Logger
	logPrefix             string
	logWriter             io.Writer
	maxIdleConns          int
	maxIdleConnsPerHost   int
	redirectFunc          func(*http.Request, []*http.Request) error
	responseHeaderTimeout time.Duration
	tlsHandshakeTimeout   time.Duration
	transport             *http.Transport
}

// Client returns the http.Client as it was initialized by the contructor, nil
// otherwise
func (c *client) Client() *http.Client {
	return c.client
}

// Close is cleanup function that makes client close any idle connections
func (c *client) Close() {
	c.log.Printf("Close client %#v with all idle connections", c)
	c.transport.CloseIdleConnections()
}

// New configures and returns a new instance of http.Client
func New(options ...Option) (Client, error) {
	c := newDefaultClient()
	if err := c.setOptions(options...); err != nil {
		return nil, err
	}
	return c, c.init()
}

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

// return a new *client with default values
func newDefaultClient() *client {
	c := new(client)
	c.setDefaults()
	return c
}

// set *sensible* default values on the *client
func (c *client) setDefaults() {
	c.dialTimeout = DefaultDialTimeout
	c.headers = make(http.Header)
	c.keepAliveTimeout = DefaultKeepAliveTimeout
	c.log = log.New(
		ioutil.Discard,
		c.logPrefix,
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC,
	)
	c.logPrefix = defaultLogPrefix
	c.maxIdleConns = DefaultMaxIdleConns
	c.maxIdleConnsPerHost = -1 //-1 means unset
	c.redirectFunc = defaultRedirectPolicy
	c.responseHeaderTimeout = DefaultResponseHeaderTimeout
	c.tlsHandshakeTimeout = DefaultTLSHandshakeTimeout
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

// initialise the client
func (c *client) init() error {
	// logger
	if c.logWriter != nil {
		c.log.SetOutput(c.logWriter)
	}
	if c.logPrefix != defaultLogPrefix {
		c.log.SetPrefix(c.logPrefix)
	}
	// if per host is unset set it to same as maxIdleConns
	if c.maxIdleConnsPerHost < 0 {
		c.maxIdleConnsPerHost = c.maxIdleConns
	}
	// create transport and client
	tr := &http.Transport{
		DialContext:           c.dialContext,
		DisableCompression:    false,
		IdleConnTimeout:       c.idleConnTimeout,
		MaxIdleConns:          c.maxIdleConns,
		MaxIdleConnsPerHost:   c.maxIdleConnsPerHost,
		ResponseHeaderTimeout: c.responseHeaderTimeout,
		TLSHandshakeTimeout:   c.tlsHandshakeTimeout,
	}
	// if disabled keep-alive say so
	if c.disableKeepAlive {
		tr.DisableKeepAlives = true
	}
	c.transport = tr
	if c.customRoundTripper == nil {
		c.customRoundTripper = tr
	}
	// disable HTTP/2
	if c.disableHTTP2 {
		nextProtoMap := make(map[string]func(string, *tls.Conn) http.RoundTripper)
		tr.TLSNextProto = nextProtoMap
	}
	c.log.Printf("initialized transport: %#v\n", tr)
	client := &http.Client{
		Transport: c.customRoundTripper,
	}
	// set redirect func
	if c.redirectFunc != nil {
		client.CheckRedirect = c.redirectFunc
	}
	c.client = client
	c.log.Printf("initialized client: %#v\n", c.client)
	return nil
}
