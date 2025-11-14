FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /service main.go

FROM alpine:3.18

COPY --from=builder /service /service

EXPOSE 8081 8082 8083 8080 9090

ENTRYPOINT ["/service"]
