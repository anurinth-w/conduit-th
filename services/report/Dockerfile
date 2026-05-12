FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum* ./
COPY shared/ shared/
COPY services/report/ services/report/
RUN cd services/report && go build -o /bin/service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8080
CMD ["./service"]
