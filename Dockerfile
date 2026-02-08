# Start from golang base image
FROM golang:1.25.7 AS builder

# Set the current working directory inside the container
WORKDIR /build

# Copy go.mod, go.sum files and download deps
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy sources to the working directory and build
COPY . .
RUN echo "Building app" && make build

# Start a new stage from debian
FROM alpine:3.22.1
LABEL org.opencontainers.image.source=https://github.com/adampresley/aletics

WORKDIR /dist

# Copy the build artifacts from the previous stage
COPY --from=builder /build/aletics .

# Run the executable
ENTRYPOINT ["./aletics"]
