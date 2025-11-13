FROM golang:1.24 AS builder

WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /spammer main.go

FROM alpine:3.18

COPY --from=builder /spammer /spammer

EXPOSE 7777

ENTRYPOINT ["/spammer"]
