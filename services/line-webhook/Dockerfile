FROM golang:1.23-alpine AS builder
WORKDIR /app

# copy workspace
COPY go.work go.work.sum* ./

# copy shared
COPY shared/ shared/

# copy go.mod และ go.sum ของทุก service
COPY services/auth/go.mod services/auth/go.mod
COPY services/auth/go.sum* services/auth/
COPY services/job/go.mod services/job/go.mod
COPY services/job/go.sum* services/job/
COPY services/user/go.mod services/user/go.mod
COPY services/user/go.sum* services/user/
COPY services/document/go.mod services/document/go.mod
COPY services/document/go.sum* services/document/
COPY services/media/go.mod services/media/go.mod
COPY services/media/go.sum* services/media/
COPY services/material/go.mod services/material/go.mod
COPY services/material/go.sum* services/material/
COPY services/notify/go.mod services/notify/go.mod
COPY services/notify/go.sum* services/notify/
COPY services/report/go.mod services/report/go.mod
COPY services/report/go.sum* services/report/
COPY services/config/go.mod services/config/go.mod
COPY services/config/go.sum* services/config/
COPY services/line-webhook/go.mod services/line-webhook/go.mod
COPY services/line-webhook/go.sum* services/line-webhook/
COPY services/ai-worker/go.mod services/ai-worker/go.mod
COPY services/ai-worker/go.sum* services/ai-worker/
COPY services/gateway/go.mod services/gateway/go.mod
COPY services/gateway/go.sum* services/gateway/

# download dependencies
RUN go work sync

# copy source เฉพาะ service นี้
COPY services/line-webhook/ services/line-webhook/

# build
RUN go build -o /bin/service ./services/line-webhook/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8080
CMD ["./service"]
