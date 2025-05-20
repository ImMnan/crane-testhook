# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

# Build the Go binary
RUN go build -o cranetest .

# Final image
FROM alpine:latest

WORKDIR /app
COPY . .

# Set execute permissions (optional, usually not needed)
# RUN chmod +x /usr/local/bin/cranetest
CMD ["/usr/local/bin/cranetest"]
