FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared ./shared
COPY habits-service ./habits-service
WORKDIR /workspace/habits-service
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /workspace/habits-service/habits-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /workspace/habits-service/habits-service .
COPY habits-service/punishments.json .
CMD ["./habits-service"]
