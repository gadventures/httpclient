package httpclient

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestClient(t *testing.T) {
	headers := make(http.Header)
	headers.Add("X-Test", "TestClient")
	h2c, err := New(
		DialTimeout(3*time.Second),
		Headers(headers),
		IdleConnTimeout(30*time.Second),
		KeepAliveTimeout(60*time.Second),
		Logger(ioutil.Discard),
		MaxIdleConns(4),
		MaxIdleConnsPerHost(2),
		RedirectPolicy(defaultRedirectPolicy),
		ResponseHeaderTimeout(20*time.Second),
		TLSHandshakeTimeout(10*time.Second),
	)
	if err != nil {
		t.Errorf("trouble when creating the client: %v", err)
	}

	tests := []struct {
		opt    func(c *client) error
		errstr string
	}{
		{badConfigOption(), "badd"},
		{MaxIdleConns(-1), ErrInvalidOptionValue.Error()},
		{MaxIdleConnsPerHost(-2), ErrInvalidOptionValue.Error()},
	}
	for _, test := range tests {
		_, err := New(test.opt)
		if err.Error() != test.errstr {
			t.Errorf("expected %s but got %s", test.errstr, err.Error())
		}
	}

	if h2c.Client() == nil {
		t.Errorf("Expected non nil *http.Client")
	}
}

func badConfigOption() func(c *client) error {
	return func(c *client) error {
		return errors.New("badd")
	}
}

func TestWithTracingOption(t *testing.T) {
	// Create an OpenTelemetry tracer, and ensure that trace propagation is enabled
	spanExporter := tracetest.NewInMemoryExporter()
	spanProcessor := trace.NewSimpleSpanProcessor(spanExporter)
	provider := trace.NewTracerProvider(trace.WithSpanProcessor(spanProcessor))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
	})

	// Create a test server for the client to talk to. Its handler will check
	// that requests from the instrumented client contain W3C Trace Context
	// propagation headers.
	server := httptest.NewServer(
		http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			if h := r.Header.Get("Traceparent"); h == "" {
				t.Errorf("expected TraceContext propagation header in request")
			}
		}),
	)
	t.Cleanup(server.Close)

	// Create a Client that is instrumented for tracing with OpenTelemetry
	client, err := New(WithTracing())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	// Make a request to the test server.
	if err := client.Get(context.Background(), NoopResponseHandler, server.URL); err != nil {
		t.Fatalf("make request: %v", err)
	}

	// Check that a client span was recorded, and that it contains the expected attribute
	spans := spanExporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected to have one span from the client, got %d: %v", len(spans), spans)
	}

	var found bool
	for _, a := range spans[0].Attributes {
		if a.Key == "http.url" {
			found = true
			if val := a.Value.AsString(); val != server.URL {
				t.Fatalf(`unexpected value for "http.url" attribute: %s`, val)
			}
			break
		}
	}
	if !found {
		t.Fatalf(`missing "http.url" attribute in span: %+v`, spans[0].Attributes)
	}
}
