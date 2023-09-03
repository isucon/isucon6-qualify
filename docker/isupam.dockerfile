FROM golang:1.21-alpine AS builder

RUN apk update

WORKDIR /app
COPY ./cmd/isupam /app

RUN GOOS=linux go build -o isupam /app/main.go

FROM alpine:3.4 AS isupam
RUN apk update

WORKDIR /app
COPY --from=builder /app/isupam /app/isupam
EXPOSE 5050

CMD ["/app/isupam"]
