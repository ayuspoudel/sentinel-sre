FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o sentinel-agent ./cmd/sentinel-agent

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/sentinel-agent /sentinel-agent

ENTRYPOINT ["/sentinel-agent"]
