FROM python:3.8

RUN pip install grpcio grpcio-tools
RUN git clone https://github.com/johananl/opentelemetry-python.git \
    && cd opentelemetry-python \
    && git checkout grpc-again \
    && pip install ext/opentelemetry-ext-grpc \
    && cd .. \
    && rm -rf opentelemetry-python

COPY . /code
WORKDIR /code

EXPOSE 9091
CMD ["python", "-u",  "field.py"]
