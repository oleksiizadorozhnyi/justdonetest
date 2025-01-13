FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

FROM alpine:latest

WORKDIR /root/

RUN apk add --no-cache bash postgresql-client

COPY --from=builder /app/app .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY ./migrations /migrations

COPY entrypoint.sh .

RUN chmod +x /root/entrypoint.sh

EXPOSE 8080

CMD ["./entrypoint.sh"]