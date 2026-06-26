FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared ./shared
COPY groups-service ./groups-service
WORKDIR /workspace/groups-service
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /workspace/groups-service/groups-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /workspace/groups-service/groups-service .
CMD ["./groups-service"]
