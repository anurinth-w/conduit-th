FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum* ./
COPY vendor/ vendor/
COPY shared/ shared/
COPY services/line-webhook/ services/line-webhook/

RUN go build -mod=vendor -o /bin/service ./services/line-webhook/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8080
CMD ["./service"]
