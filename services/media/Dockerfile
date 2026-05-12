FROM golang:1.23-alpine AS builder
WORKDIR /app

# copy workspace และ shared ก่อน
COPY go.work ./
COPY shared/ shared/

# copy เฉพาะ service นี้
COPY services/media/ services/media/

# build จาก root workspace
RUN go build -o /bin/service ./services/media/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8080
CMD ["./service"]
