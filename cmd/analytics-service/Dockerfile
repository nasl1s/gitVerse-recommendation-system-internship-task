FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./configs /root/configs

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o analytics-service ./cmd/analytics-service/main.go

FROM alpine:3.18

WORKDIR /app

RUN mkdir -p /app/logger

COPY --from=builder /app/analytics-service .

VOLUME /app/logger

RUN chmod +x ./analytics-service

CMD ["./analytics-service"]