FROM golang:1.16-alpine AS builder
RUN apk add --no-cache \
    ca-certificates \
    git

RUN mkdir /usr/local/redplant
WORKDIR /usr/local/redplant

COPY . .

RUN go get
RUN go build -o redplant *.go

FROM alpine:3.12.4
RUN mkdir /usr/local/redplant
WORKDIR /usr/local/redplant
COPY --from=builder /usr/local/redplant/redplant .

RUN addgroup -g 1000 redplant && \
    adduser -h /usr/local/redplant -D -u 1000 -G redplant redplant && \
    chown -R redplant:redplant /usr/local/redplant

USER redplant

ENTRYPOINT [ "/usr/local/redplant/redplant" ]
CMD ["-c", "etc/config.yaml", "-l", "etc/logging.yaml"]