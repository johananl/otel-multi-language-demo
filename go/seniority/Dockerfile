FROM golang as builder

COPY . /code
WORKDIR /code
ENV GOPROXY=https://proxy.golang.org
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./seniority ./cmd

FROM scratch

COPY --from=builder /code/seniority /
EXPOSE 9090
CMD ["/seniority"]
