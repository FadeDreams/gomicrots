FROM golang:1.21.3-alpine3.18

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

# Download Go module dependencies
RUN go mod download

# Copy only the necessary files into the container
COPY ./cmd/web /app/cmd/web
COPY ./cmd/web/templates /app/cmd/web/templates

# Build the Go app
RUN go build -o ./app ./cmd/web

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the binary program produced by `go build`
CMD ["./app"]
