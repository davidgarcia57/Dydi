FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /app/realtime-service ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/realtime-service .
# CMD not ENTRYPOINT — allows Render SIGTERM to reach the Go process directly
CMD ["./realtime-service"]
