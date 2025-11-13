FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /pingr .

FROM alpine:3.18

WORKDIR /app

RUN apk add --update --no-cache graphviz

COPY --from=builder /pingr .

COPY config/config.yaml ./config/config.yaml

ENTRYPOINT ["./pingr", "run"]

CMD ["-c", "/app/config/config.yaml"]
