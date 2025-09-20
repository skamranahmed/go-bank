# Stage 1: Build the Go binary
FROM golang:1.25.1-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o go-bank-api .

# Stage 2: Run the binary in a minimal image
FROM alpine:3.21

WORKDIR /root

COPY ./config ./config

COPY --from=builder /app/go-bank-api .

EXPOSE 8080

CMD ["./go-bank-api"]