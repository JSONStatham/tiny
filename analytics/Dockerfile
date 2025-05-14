FROM golang:1.24-alpine

RUN apk update && apk add bash ca-certificates git curl gcc g++ libc-dev librdkafka-dev pkgconf

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -tags musl -v -o /app/bin/analytics ./cmd/analytics/main.go

CMD ["/app/bin/analytics"]
