# ---------- build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Required for HTTPS (go modules, Postgres certs if needed later)
RUN apk add --no-cache ca-certificates

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy entire repository (build context is repo root)
COPY . .

# Build the cluster registry binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o sentinel-cluster-registry ./cmd/sentinel-cluster-registry


# ---------- runtime stage ----------
FROM gcr.io/distroless/base-debian12

WORKDIR /

# Copy binary from builder
COPY --from=builder /app/sentinel-cluster-registry /sentinel-cluster-registry

# Run as non-root
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/sentinel-cluster-registry"]
