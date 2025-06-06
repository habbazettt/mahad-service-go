# Stage 1: Builder
FROM golang:1.24-alpine AS builder

ENV GIN_MODE=release \
    GO111MODULE=on \
    GOPATH=/go

RUN apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Stage 2: Final Image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

COPY q_table_model.json ./
COPY historical_best.json ./

EXPOSE 8080

ENTRYPOINT ["/root/main"]