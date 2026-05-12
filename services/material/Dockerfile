FROM golang:1.23-alpine AS builder
WORKDIR /app

# copy go.work
COPY go.work go.work.sum* ./

# copy shared
COPY shared/ shared/

# copy go.mod ของทุก service (go.work ต้องการ)
COPY services/auth/go.mod services/auth/go.mod
COPY services/job/go.mod services/job/go.mod
COPY services/user/go.mod services/user/go.mod
COPY services/document/go.mod services/document/go.mod
COPY services/media/go.mod services/media/go.mod
COPY services/material/go.mod services/material/go.mod
COPY services/notify/go.mod services/notify/go.mod
COPY services/report/go.mod services/report/go.mod
COPY services/config/go.mod services/config/go.mod
COPY services/line-webhook/go.mod services/line-webhook/go.mod
COPY services/ai-worker/go.mod services/ai-worker/go.mod
COPY services/gateway/go.mod services/gateway/go.mod

# copy source เฉพาะ service นี้
COPY services/material/ services/material/

# build
RUN go build -o /bin/service ./services/material/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8080
CMD ["./service"]
