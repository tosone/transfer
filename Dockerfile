FROM golang:alpine AS BUILD

ARG PKG=/go/src/transfer

WORKDIR ${PKG}

ADD . ${PKG}

RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache make build-base && make release && cp ./release/transfer /tmp

FROM alpine:3.12

WORKDIR /app

COPY --from=BUILD /tmp/transfer /usr/local/bin
COPY config.yaml /etc/transfer/

VOLUME /app

EXPOSE 3000

CMD ["transfer"]
