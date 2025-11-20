# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with tview tag
RUN go build -tags tview -ldflags="-s -w" -o lazygotest ./cmd/lazygotest

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/lazygotest /usr/local/bin/lazygotest

# Set entrypoint
ENTRYPOINT ["lazygotest"]

# Default to current directory
CMD ["."]
