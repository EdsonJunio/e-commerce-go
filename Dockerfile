# --- Build Stage ---
FROM golang:1.25.1-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary named 'main'
# CGO_ENABLED=0 -> Static binary (no external C dependencies)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

# --- Final Stage ---
FROM scratch

WORKDIR /app

# Copy necessary system files from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/main .

# Set non-root user
USER 65532:65532

# Expose port
EXPOSE 8081

# Run!
ENTRYPOINT ["./main"]