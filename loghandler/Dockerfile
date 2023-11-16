
from golang:1.21.3-alpine3.18

# set the current working directory inside the container
workdir /app

# we want to populate the module cache based on the go.{mod,sum} files.
copy go.mod .
copy go.sum .

# download go module dependencies
run go mod download

# copy only the necessary files into the container
copy ./cmd/api /app/cmd/api

# build the go app
run go build -o ./app ./cmd/api

# this container exposes port 8082 to the outside world
expose 8082

# run the binary program produced by `go build`
cmd ["./app"]

