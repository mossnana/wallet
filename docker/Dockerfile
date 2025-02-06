FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./wallet-api ./cmd/api/main.go

FROM alpine:3.21
WORKDIR /root/
COPY --from=builder /app/wallet-api .
EXPOSE 3000
CMD ["./wallet-api"]
