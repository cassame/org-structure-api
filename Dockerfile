FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api-server ./cmd/api/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/api-server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./api-server"]