FROM golang:1.22 AS build-stage

WORKDIR /app

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

FROM build-stage AS builder
WORKDIR /app
COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o clamd_exporter ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/clamd_exporter .

EXPOSE 8181

ENTRYPOINT ["./clamd_exporter"]
CMD ["-l", ":8181","-a", "192.168.127.131:3310", "-n", "tcp"]
