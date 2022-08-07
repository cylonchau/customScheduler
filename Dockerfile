FROM golang:alpine AS builder
MAINTAINER cylon
WORKDIR /scheduler
COPY ./ /scheduler
ENV GOPROXY https://goproxy.cn,direct
RUN \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add upx  && \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o scheduler main.go && \
    upx -1 scheduler && \
    chmod +x scheduler

FROM alpine AS runner
WORKDIR /go/scheduler
COPY --from=builder /scheduler/scheduler .
COPY --from=builder /scheduler/scheduler.yaml /etc/
VOLUME ["./scheduler"]