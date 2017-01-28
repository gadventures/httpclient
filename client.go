package httpclient

import (
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
	defaultLogPrefix        = "httpclient "
)

//ErrInvalidValue signifies that an invaild value was given to configartion option
var ErrInvalidValue = errors.New("invalid value for option")

// Client is our http client that can be used to make http requests efficiently
// for now we intend to support Get
// it is used rather than net/http package directly due to ability to configure
// timeouts and watch the resource use
// safe (and intended) to use from several go routines
type Client struct {
	headers               http.Header
	redirectFunc          func(*http.Request, []*http.Request) error
	client                *http.Client
	maxIdleConns          int
	maxIdleConnsPerHost   int
	logWriter             io.Writer
	log                   *log.Logger
	currentConnID         int64
	transport             *http.Transport
	dialTimeout           time.Duration
	keepAliveTimeout      time.Duration
	idleConnTimeout       time.Duration
	tlsHandshakeTimeout   time.Duration
	responseHeaderTimeout time.Duration
	logPrefix             string
}

// Close is cleanup function that makes client
// close any idle connections
func (c *Client) Close() {
	c.log.Printf("Close client  %#v with all idle connections", c)
	c.transport.CloseIdleConnections()
}

// New configures and returns a new instance of http.Client
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
// use this for headers shared among all requests
// for request specific headers use the ExtraHeaders RequestOption
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
// the default behaviour is to copy the headers from original
// request and try again up to 10 times
func RedirectPolicy(redirectFunc func(req *http.Request, via []*http.Request) error) func(*Client) error {
	return func(c *Client) error {
		c.redirectFunc = redirectFunc
		return nil
	}
}

// DialTimeout is configuration option to pass to Client
// it changes how long will the client wait to establish
// the TCP connection
func DialTimeout(t time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.dialTimeout = t
		return nil
	}
}

// IdleConnTimeout is a configuration option to pass to Client
// it sets how long will idle connections live in a pool
// waiting to be used again.
// Once the timeout is reached connections are closed and removed from pool.
func IdleConnTimeout(t time.Duration) func(*Client) error {
	return func(c *Client) error {
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
func KeepAliveTimeout(t time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.keepAliveTimeout = t
		return nil
	}
}

// TLSHandshakeTimeout is a configuration option to pass to Client
// it limits the time spent performing the TLS handshake.
func TLSHandshakeTimeout(t time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.tlsHandshakeTimeout = t
		return nil
	}
}

// ResponseHeaderTimeout is a configuration option to pass to Client
// it limits the time spent reading the headers of the response.
func ResponseHeaderTimeout(t time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.responseHeaderTimeout = t
		return nil
	}
}

// MaxIdleConns is configuration option to pass to Client
// it changes the maximum idle connections that the client will keep
// in a pool
func MaxIdleConns(n int) func(*Client) error {
	return func(c *Client) error {
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
func MaxIdleConnsPerHost(n int) func(*Client) error {
	return func(c *Client) error {
		if n < 0 {
			return ErrInvalidValue
		}
		c.maxIdleConnsPerHost = n
		return nil
	}
}

// Logger is configuration option to pass to Client
// it changes where the debug info is written
// (ioutil.Discard by default)
func Logger(w io.Writer) func(*Client) error {
	return func(c *Client) error {
		c.logWriter = w
		return nil
	}
}

// LogPrefix is configuration option to pass to Client
// to change the prefix used in Clients log output.
// This can be useful when one is using several httpclients.
func LogPrefix(p string) func(*Client) error {
	return func(c *Client) error {
		c.logPrefix = p
		return nil
	}
}

func (c *Client) setDefaults() {
	c.headers = make(http.Header)
	c.dialTimeout = DefaultDialTimeout
	c.keepAliveTimeout = DefaultKeepAliveTimeout
	c.tlsHandshakeTimeout = DefaultTLSHandshakeTimeout
	c.responseHeaderTimeout = DefaultResponseHeaderTimeout
	c.maxIdleConns = DefaultMaxIdleConns
	c.maxIdleConnsPerHost = -1 //-1 means unset
	c.redirectFunc = defaultRedirectPolicy
	c.logPrefix = defaultLogPrefix
	c.log = log.New(
		ioutil.Discard,
		c.logPrefix,
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
}

func (c *Client) setUp() error {
	//logger
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
		MaxIdleConns:          c.maxIdleConns,
		MaxIdleConnsPerHost:   c.maxIdleConnsPerHost,
		DisableCompression:    false,
		DialContext:           c.dialContext,
		TLSHandshakeTimeout:   c.tlsHandshakeTimeout,
		ResponseHeaderTimeout: c.responseHeaderTimeout,
		IdleConnTimeout:       c.idleConnTimeout,
	}
	c.transport = tr
	c.log.Printf("Initialized transport: %#v\n", tr)
	client := &http.Client{
		Transport: tr,
	}
	//set redirect func
	if c.redirectFunc != nil {
		client.CheckRedirect = c.redirectFunc
	}
	c.client = client
	c.log.Printf("Initialized client: %#v\n", c.client)
	return nil
}
