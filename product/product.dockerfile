# Base Go Image
FROM golang:1.24.0-alpine AS builder

# Set working directory
WORKDIR /app

# Add source code
COPY . /app

# Build the binary and add environment variable through CGO_ENABLED
RUN CGO_ENABLED=0 go build -o product ./cmd/api

RUN chmod +x /app/product

# Build a small image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder app/product ./

# Copy migrations files
COPY migrations ./migrations

# Expose HTTP port
EXPOSE 8080

# Command to run the executable
CMD ["./product"]
