FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o main ./cmd/server/main.go

FROM alpine:3.18
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]