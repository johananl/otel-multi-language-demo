package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/johananl/otel-multi-language-demo/go/role/pkg/tracing"
	pb "github.com/johananl/otel-multi-language-demo/go/role/proto"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

var roles []string = []string{
	"coordinator",
	"manager",
	"trainer",
	"dictator",
	"tamer",
	"analyst",
	"engineer",
	"evangelist",
	"designer",
	"plumber",
	"consultant",
	"optimizer",
	"specialist",
	"researcher",
	"scientist",
}

type server struct {
	pb.UnimplementedRoleServer
}

func (s *server) GetRole(ctx context.Context, in *pb.RoleRequest) (*pb.RoleReply, error) {
	log.Println("Received role request")

	// Get current span. The span was created within the gRPC interceptor.
	// We are retrieving it here because we want to add data to it.
	span := trace.SpanFromContext(ctx)

	if in.Slow {
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	}
	if in.Unreliable {
		// Return an error 10% of the time.
		if rand.Intn(10) == 0 {
			// Mark the span as containing an error.
			span.SetStatus(500, "Random error")
			return nil, errors.New("random error")
		}

	}
	selected := roles[rand.Intn(len(roles))]

	// Log the result on the span.
	span.AddEvent(ctx, "Selected role", key.New("role").String(selected))

	return &pb.RoleReply{Role: selected}, nil
}

func initTraceProvider(jaegerHost, jaegerPort string) func() {
	_, flush, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(fmt.Sprintf("http://%s:%s/api/traces", jaegerHost, jaegerPort)),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "role",
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

	host := getenv("ROLE_HOST", "localhost")
	port := getenv("ROLE_PORT", "9092")

	jaegerHost := getenv("ROLE_JAEGER_HOST", "localhost")
	jaegerPort := getenv("ROLE_JAEGER_PORT", "14268")

	// Initialize tracing.
	fn := initTraceProvider(jaegerHost, jaegerPort)
	defer fn()

	addr := fmt.Sprintf("%s:%s", host, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(tracing.UnaryServerInterceptor))
	pb.RegisterRoleServer(s, &server{})

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
