# Use the official Go image as the base image
FROM golang:1.17-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Install build dependencies
RUN apk --no-cache add build-base

# Copy the Go module files to the container
COPY go.mod go.sum ./

# Download and cache the Go module dependencies
RUN go mod download

# Copy the Go application source code to the container
COPY main.go ./

# Build the Go application inside the container
RUN CGO_ENABLED=1 go build -o kafka-app

# Create a new Docker image with a minimal Alpine base image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the build stage
COPY --from=build /app/kafka-app ./

# Run the Kafka application
CMD ["./kafka-app"]
