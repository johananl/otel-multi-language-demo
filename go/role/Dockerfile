FROM golang as builder

COPY . /code
WORKDIR /code
ENV GOPROXY=https://proxy.golang.org
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./role ./cmd

FROM scratch

COPY --from=builder /code/role /
EXPOSE 9092
CMD ["/role"]
