FROM golang:alpine AS BUILD

ARG PKG=/go/src/transfer

WORKDIR ${PKG}

ADD . ${PKG}

RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache make && make release && cp ./release/transfer /tmp

FROM alpine:3.12

COPY --from=BUILD /tmp/transfer /usr/local/bin

EXPOSE 3000

CMD ["transfer"]
