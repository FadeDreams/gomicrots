
FROM golang:1.21.3-alpine3.18

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

# Download Go module dependencies
RUN go mod download

# Copy only the necessary files into the container
COPY ./cmd/api /app/cmd/api
COPY ./models /app/models

# Build the Go app
RUN go build -o ./app ./cmd/api

# This container exposes port 8082 to the outside world
EXPOSE 8082

# Run the binary program produced by `go build`
CMD ["./app"]
