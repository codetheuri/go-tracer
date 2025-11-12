# Step 1: Build the binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install libpcap dependencies
RUN apk add --no-cache libpcap-dev gcc musl-dev

# Copy module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 go build -o /network-agent ./cmd/agent/main.go

# Step 2: Create the final lightweight image
FROM alpine:latest

WORKDIR /

# Install only libpcap (runtime)
RUN apk add --no-cache libpcap

# Copy the binary from the builder stage
COPY --from=builder /network-agent /network-agent

# Run the binary
CMD ["/network-agent"]