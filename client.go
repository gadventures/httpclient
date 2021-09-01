package httpclient

import (
	"context"
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

// ErrInvalidOptionValue is returned when an invalid value was given for some
// configuration option
var ErrInvalidOptionValue = errors.New("invalid value for option")

type Client interface {
	// Client returns the underlying *http.Client
	Client() *http.Client

	// Close all Idle connections
	Close()

	// HTTP Request Methods //

	// Do is the generic HTTP request method. The two string parameters in
	// order are the HTTP Method, and the URL to request respectively
	Do(ctx context.Context, rh ResponseHandler, method, url string, body io.Reader, opts ...RequestOption) error

	// GET method
	Get(ctx context.Context, rh ResponseHandler, url string, opts ...RequestOption) error

	// POST method
	Post(ctx context.Context, rh ResponseHandler, url string, body io.Reader, opts ...RequestOption) error
}

// ensure Client interface implementation
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

// Client returns the *http.Client as it was initialized by the constructor
func (c *client) Client() *http.Client {
	return c.client
}

// Close is a cleanup function that makes the client close any idle connections
func (c *client) Close() {
	c.log.Printf("Close client %#v with all idle connections", c)
	c.transport.CloseIdleConnections()
}

// New configures and returns a new instance of Client
func New(options ...Option) (Client, error) {
	c := newDefaultClient()
	if err := c.setOptions(options...); err != nil {
		return nil, err
	}
	return c, c.init()
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
	// create transport
	tr := &http.Transport{
		DialContext:           c.dialContext,
		DisableCompression:    false,
		DisableKeepAlives:     c.disableKeepAlive,
		IdleConnTimeout:       c.idleConnTimeout,
		MaxIdleConns:          c.maxIdleConns,
		MaxIdleConnsPerHost:   c.maxIdleConnsPerHost,
		ResponseHeaderTimeout: c.responseHeaderTimeout,
		TLSHandshakeTimeout:   c.tlsHandshakeTimeout,
		ForceAttemptHTTP2:     !c.disableHTTP2,
	}
	c.transport = tr
	if c.customRoundTripper == nil {
		c.customRoundTripper = tr
	}
	c.log.Printf("initialized transport: %#v\n", tr)
	// create client
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
