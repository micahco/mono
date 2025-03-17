FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate

FROM alpine:3.19

# Add necessary system dependencies (PostgreSQL client for potential troubleshooting)
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

COPY --from=builder /migrate .

ENTRYPOINT ["./migrate"]