FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/user-service

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]