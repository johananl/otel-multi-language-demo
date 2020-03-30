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

	"github.com/johananl/otel-multi-language-demo/go/seniority/pkg/tracing"
	pb "github.com/johananl/otel-multi-language-demo/go/seniority/proto"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

var seniorities []string = []string{
	"senior",
	"junior",
	"assistant",
	"executive",
	"intergalactic",
	"lead",
	"corporate",
	"regional",
	"principal",
	"chief",
}

type server struct {
	pb.UnimplementedSeniorityServer
}

func (s *server) GetSeniority(ctx context.Context, in *pb.SeniorityRequest) (*pb.SeniorityReply, error) {
	log.Println("Received seniority request")

	// Get current span. The span was created within the gRPC interceptor.
	// We are just adding data to it here.
	span := trace.SpanFromContext(ctx)

	if in.Slow {
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	}
	if in.Unreliable {
		// Return an error 10% of the time.
		if rand.Intn(10) == 0 {
			span.SetStatus(500, "Random error")
			return nil, errors.New("random error")
		}

	}
	selected := seniorities[rand.Intn(len(seniorities))]

	span.AddEvent(ctx, "Selected seniority", key.New("seniority").String(selected))

	return &pb.SeniorityReply{Seniority: selected}, nil
}

func initTraceProvider(jaegerHost, jaegerPort string) func() {
	_, flush, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(fmt.Sprintf("http://%s:%s/api/traces", jaegerHost, jaegerPort)),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "seniority",
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

	host := getenv("SENIORITY_HOST", "localhost")
	port := getenv("SENIORITY_PORT", "9090")

	jaegerHost := getenv("SENIORITY_JAEGER_HOST", "localhost")
	jaegerPort := getenv("SENIORITY_JAEGER_PORT", "14268")

	fn := initTraceProvider(jaegerHost, jaegerPort)
	defer fn()

	addr := fmt.Sprintf("%s:%s", host, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(tracing.UnaryServerInterceptor))
	pb.RegisterSeniorityServer(s, &server{})

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
