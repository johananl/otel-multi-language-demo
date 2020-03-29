from concurrent import futures
import logging

import grpc

import field_pb2
import field_pb2_grpc


class Field(field_pb2_grpc.FieldServicer):

    def GetField(self, request, context):
        print('Received field request')
        return field_pb2.FieldReply()


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    # helloworld_pb2_grpc.add_GreeterServicer_to_server(Greeter(), server)
    field_pb2_grpc.add_FieldServicer_to_server(Field(), server)
    server.add_insecure_port('[::]:9091')
    server.start()
    print('Listening for gRPC connections on port 9091')
    server.wait_for_termination()


if __name__ == '__main__':
    logging.basicConfig()
    serve()
