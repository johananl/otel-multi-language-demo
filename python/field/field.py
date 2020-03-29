from concurrent import futures
import logging
import random

import grpc

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

class Field(field_pb2_grpc.FieldServicer):

    def GetField(self, request, context):
        print('Received field request')
        selected = fields[random.randint(0, len(fields)-1)]
        return field_pb2.FieldReply(field=selected)


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    field_pb2_grpc.add_FieldServicer_to_server(Field(), server)
    server.add_insecure_port('[::]:9091')
    server.start()
    print('Listening for gRPC connections on port 9091')
    server.wait_for_termination()


if __name__ == '__main__':
    logging.basicConfig()
    serve()
