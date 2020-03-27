module github.com/johananl/otel-multi-language-demo/go/frontend

go 1.14

require (
	github.com/johananl/otel-multi-language-demo/go/field v0.0.0-00010101000000-000000000000
	github.com/johananl/otel-multi-language-demo/go/role v0.0.0-00010101000000-000000000000
	github.com/johananl/otel-multi-language-demo/go/seniority v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v0.2.2-0.20200111012159-d85178b63b15
	go.opentelemetry.io/otel/exporter/trace/jaeger v0.2.2-0.20200111012159-d85178b63b15
	google.golang.org/grpc v1.24.0
)

replace (
	github.com/johananl/otel-multi-language-demo/go/field => ../field
	github.com/johananl/otel-multi-language-demo/go/role => ../role
	github.com/johananl/otel-multi-language-demo/go/seniority => ../seniority
)
