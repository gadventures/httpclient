module github.com/gadventures/httpclient

go 1.16

retract (
	v1.0.0-rc.1
	v0.2.0-rc.2
	v0.2.0-rc.1
)

require (
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
)
