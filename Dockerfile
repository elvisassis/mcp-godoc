# Use a lightweight Go image for building the application
FROM golang:1.25-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Copy the vendored dependencies
COPY vendor ./vendor

# Copy the rest of the source code
COPY cmd ./cmd
COPY internal ./internal

# Build the application using the vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/godoctor ./cmd/godoctor   

# Use a minimal base image for the final image
FROM alpine:latest

# Copy the built executable from the builder stage
COPY --from=builder /app/godoctor /usr/local/bin/godoctor

# Set the command to run the godoctor executable
CMD ["godoctor"]
