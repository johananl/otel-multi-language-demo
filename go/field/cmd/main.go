package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/johananl/otel-multi-language-demo/go/field/pkg/tracing"
	pb "github.com/johananl/otel-multi-language-demo/go/field/proto"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

var fields []string = []string{
	"marketing",
	"dolphin",
	"cat",
	"penguin",
	"engineering",
	"aerospace",
	"machinery",
	"finance",
	"strategy",
	"beer",
	"coffee",
	"whisky",
	"laundry",
	"socks",
}

type server struct {
	pb.UnimplementedFieldServer
}

func (s *server) GetField(ctx context.Context, in *pb.FieldRequest) (*pb.FieldReply, error) {
	log.Println("Received field request")

	if in.Slow {
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	}
	selected := fields[rand.Intn(len(fields))]

	// Get current span. The span was created within the gRPC interceptor.
	// We are just adding data to it here.
	span := trace.SpanFromContext(ctx)
	span.AddEvent(ctx, "Selected field", key.New("field").String(selected))

	return &pb.FieldReply{Field: selected}, nil
}

func initTraceProvider(jaegerHost, jaegerPort string) func() {
	_, flush, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(fmt.Sprintf("http://%s:%s/api/traces", jaegerHost, jaegerPort)),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "field",
			Tags: []core.KeyValue{
				key.String("exporter", "jaeger"),
			},
		}),
		jaeger.RegisterAsGlobal(),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		flush()
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	host := getenv("FIELD_HOST", "localhost")
	port := getenv("FIELD_PORT", "9091")

	jaegerHost := getenv("FIELD_JAEGER_HOST", "localhost")
	jaegerPort := getenv("FIELD_JAEGER_PORT", "14268")

	fn := initTraceProvider(jaegerHost, jaegerPort)
	defer fn()

	addr := fmt.Sprintf("%s:%s", host, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(tracing.UnaryServerInterceptor))
	pb.RegisterFieldServer(s, &server{})

	ch := make(chan struct{})
	go func(ch chan struct{}) {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		ch <- struct{}{}
	}(ch)
	log.Printf("Listening for gRPC connections on port %s", port)

	<-ch
}
