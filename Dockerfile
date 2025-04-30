FROM golang:1.24-alpine AS builder

COPY . /github.com/nogavadu/load_balancer
WORKDIR /github.com/nogavadu/load_balancer

RUN go mod download
RUN go build -o ./bin/load_balancer ./cmd/load_balancer/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/nogavadu/load_balancer/bin/load_balancer .
COPY --from=builder /github.com/nogavadu/load_balancer/config/config.yaml .

ENV CONFIG_PATH="./config.yaml"

CMD ["./load_balancer"]