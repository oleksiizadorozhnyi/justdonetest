FROM golang:1.23

WORKDIR /app

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd

EXPOSE 8080

CMD ["./app"]