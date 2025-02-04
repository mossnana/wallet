FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /wallet-api

FROM alpine:3.21
WORKDIR /root/
COPY --from=builder /wallet-api .
EXPOSE WALLET_API_PORT
CMD ["./wallet-api"]
