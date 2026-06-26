FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared ./shared
COPY realtime-service ./realtime-service
WORKDIR /workspace/realtime-service
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /workspace/realtime-service/realtime-service ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /workspace/realtime-service/realtime-service .
# CMD not ENTRYPOINT — allows Render SIGTERM to reach the Go process directly
CMD ["./realtime-service"]
