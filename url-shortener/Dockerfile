FROM golang:1.24-alpine

RUN apk update && apk add bash ca-certificates git curl gcc g++ libc-dev librdkafka-dev pkgconf

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -tags musl -v -o /app/bin/url-shortener ./cmd/url-shortener/main.go

EXPOSE 8080

CMD ["/app/bin/url-shortener"]
