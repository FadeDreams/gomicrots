# Use the official Golang image
FROM golang:1.21.3-alpine3.18

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to the working directory
COPY go.mod .
COPY go.sum .

# Download Go module dependencies
RUN go mod download

# Copy the entire project to the working directory
COPY . .

# Build the Go app
RUN go build -o app ./main.go

# Add execution permissions to the app binary
RUN chmod +x app

# Expose port 8084 for the application
EXPOSE 8084

# Command to run the executable
CMD ["./app"]
