from concurrent import futures
import logging
import random
import time

import grpc

from opentelemetry import trace
from opentelemetry.trace.status import StatusCanonicalCode
from opentelemetry.ext import jaeger
from opentelemetry.ext.grpc import server_interceptor
from opentelemetry.ext.grpc.grpcext import intercept_server
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchExportSpanProcessor

import field_pb2
import field_pb2_grpc

fields = [
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
]

# Initialize tracing.
trace.set_tracer_provider(TracerProvider())
jaeger_exporter = jaeger.JaegerSpanExporter(
    service_name="field",
    agent_host_name="jaeger",
    agent_port=6831,
)
span_processor = BatchExportSpanProcessor(jaeger_exporter)
trace.get_tracer_provider().add_span_processor(span_processor)
tracer = trace.get_tracer(__name__)

class Field(field_pb2_grpc.FieldServicer):
    def GetField(self, request, context):
        log = logging.getLogger()
        log.info('Received field request')

        # Get current span. The span was created within the gRPC interceptor.
        # We are retrieving it here because we want to add data to it.
        span = tracer.get_current_span()

        if request.slow:
            time.sleep(random.randint(0, 300) / 1000)
        if request.unreliable:
            # Return an error 10% of the time.
            if random.randint(0, 10) == 0:
                # Mark the span as containing an error.
                span.set_status(
                    trace.status.Status(
                        StatusCanonicalCode.UNAVAILABLE,
                        'Random error',
                    )
                )
                context.set_code(grpc.StatusCode.UNKNOWN)
                return field_pb2.FieldReply()
        selected = fields[random.randint(0, len(fields)-1)]

        # Log the result on the span.
        span.add_event('Selected field', {'field': selected})

        return field_pb2.FieldReply(field=selected)


def serve():
    log = logging.getLogger()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    # OpenTelemetry magic!
    server = intercept_server(server, server_interceptor(tracer))

    field_pb2_grpc.add_FieldServicer_to_server(Field(), server)
    server.add_insecure_port('[::]:9091')
    server.start()
    log.info('Listening for gRPC connections on port 9091')
    server.wait_for_termination()


if __name__ == '__main__':
    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s %(message)s',
        datefmt='%Y/%m/%d %I:%M:%S',
    )
    serve()
