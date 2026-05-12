FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.work ./
COPY shared/ shared/

COPY services/auth/go.mod services/auth/go.sum* services/auth/
COPY services/job/go.mod services/job/go.sum* services/job/
COPY services/user/go.mod services/user/go.sum* services/user/
COPY services/document/go.mod services/document/go.sum* services/document/
COPY services/media/go.mod services/media/go.sum* services/media/
COPY services/material/go.mod services/material/go.sum* services/material/
COPY services/notify/go.mod services/notify/go.sum* services/notify/
COPY services/report/go.mod services/report/go.sum* services/report/
COPY services/config/go.mod services/config/go.sum* services/config/
COPY services/line-webhook/go.mod services/line-webhook/go.sum* services/line-webhook/
COPY services/ai-worker/go.mod services/ai-worker/go.sum* services/ai-worker/
COPY services/gateway/go.mod services/gateway/go.sum* services/gateway/

RUN cd services/document && go mod download

COPY services/document/ services/document/

RUN go build -o /bin/service ./services/document/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8080
CMD ["./service"]
