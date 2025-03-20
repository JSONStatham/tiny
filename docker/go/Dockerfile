FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /app/bin/tiny ./cmd/tiny/main.go


EXPOSE 8080

CMD ["/app/bin/tiny"]