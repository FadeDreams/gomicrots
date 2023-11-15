# base go image
FROM golang:1.16.13-alpine3.16 as builder

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 go build -o service1App ./cmd/api
RUN chmod +x /app/service1App

# build a tiny docker image
FROM alpine:3.16

RUN mkdir /app
COPY --from=builder /app/service1App /app

CMD [ "/app/service1App" ]

