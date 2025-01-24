FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o sso-service ./cmd/sso-service/main.go

FROM alpine:3.18

WORKDIR /app

RUN mkdir -p /app/logger

COPY --from=builder /app/sso-service .

VOLUME /app/logger

RUN chmod +x ./sso-service

CMD ["./sso-service"]