FROM python:3.8

RUN pip install grpcio grpcio-tools

COPY . /code
WORKDIR /code

EXPOSE 9091
CMD ["python",  "field.py"]
