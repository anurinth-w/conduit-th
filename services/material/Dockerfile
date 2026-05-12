FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum* ./
COPY vendor/ vendor/
COPY shared/ shared/
COPY services/material/ services/material/

RUN go build -mod=vendor -o /bin/service ./services/material/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8080
CMD ["./service"]
