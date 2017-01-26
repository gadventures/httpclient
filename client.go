package httpclient

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	// DefaultTimeout for network exchange
	DefaultTimeout = 60 * time.Second
	// DefaultMaxIdleConns to keep
	DefaultMaxIdleConns = 15
)

// Client is our http client that can be used to make http requests efficiently
// for now we intend to support Get
// it is used rather than net/http package directly due to ability to configure
// timeouts and watch the resource use
// safe (and intended) to use from several go routines
type Client struct {
	headers          http.Header
	redirectFunc     func(*http.Request, []*http.Request) error
	client           *http.Client
	maxIdleConns     int
	logWriter        io.Writer
	log              *log.Logger
	currentConnID    int64
	transport        *http.Transport
	dialTimeout      time.Duration
	keepAliveTimeout time.Duration
}

// Close forces client to close any idle connections and cleanup
func (c *Client) Close() {
	c.log.Printf("Close client  %#v with all idle connections", c)
	c.transport.CloseIdleConnections()
}

// New configures and returns new instance of Client
func New(options ...func(*Client) error) (*Client, error) {
	c := new(Client)
	c.setDefaults()
	for _, opt := range options {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return c, err
		}
	}
	err := c.setUp()
	return c, err
}

// Headers is configuration option to pass headers to Client
// it makes GET requests use the provided headers
func Headers(headers http.Header) func(*Client) error {
	return func(c *Client) error {
		for k, v := range headers {
			c.headers[k] = v
		}
		return nil
	}
}

// RedirectPolicy is configuration option to pass to Client
// it changes what client does on redirects
func RedirectPolicy(redirectFunc func(req *http.Request, via []*http.Request) error) func(*Client) error {
	return func(c *Client) error {
		c.redirectFunc = redirectFunc
		return nil
	}
}

// DialTimeout is configuration option to pass to Client
// it changes the timeout for request exchange
func DialTimeout(t time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.dialTimeout = t
		return nil
	}
}

// KeepAliveTimeout is configuration option to pass to Client
// it changes the timeout for how long the tcp connection stays open
func KeepAliveTimeout(t time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.keepAliveTimeout = t
		return nil
	}
}

// MaxIdleConns is configuration option to pass to Client
// it changes the maximum idle connections that the client will keep around
func MaxIdleConns(n int) func(*Client) error {
	return func(c *Client) error {
		c.maxIdleConns = n
		return nil
	}
}

// Logger is configuration option to pass to Client
// it changes where the debug info is written stderr by default
func Logger(w io.Writer) func(*Client) error {
	return func(c *Client) error {
		c.logWriter = w
		return nil
	}
}

func (c *Client) setDefaults() {
	c.headers = make(http.Header)
	c.dialTimeout = DefaultTimeout
	c.keepAliveTimeout = DefaultTimeout
	c.maxIdleConns = DefaultMaxIdleConns
	c.redirectFunc = defaultRedirectPolicy
	c.log = log.New(
		ioutil.Discard,
		"httpclient ",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
}

func (c *Client) setUp() error {
	//logger
	if c.logWriter != nil {
		c.log.SetOutput(c.logWriter)
	}
	//create transport and client
	tr := &http.Transport{
		MaxIdleConns:       c.maxIdleConns,
		DisableCompression: false,
		DialContext:        c.dialContext,
	}
	c.transport = tr
	c.log.Printf("Initialized transport: %#v\n", tr)
	client := &http.Client{
		Transport: tr,
		Timeout:   c.keepAliveTimeout,
	}
	//set redirect func
	if c.redirectFunc != nil {
		client.CheckRedirect = c.redirectFunc
	}
	c.client = client
	c.log.Printf("Initialized client: %#v\n", c.client)
	return nil
}
