FROM golang:1.17.0-alpine as server-build

WORKDIR /github.com/traPtitech/knoq

COPY go.mod go.sum ./
ENV GO111MODULE=on
RUN go mod download
COPY ./ ./

RUN go build -o knoq

FROM alpine:3.13.5

WORKDIR /app

ENV DOCKERIZE_VERSION v0.6.1

RUN apk --update add tzdata \
  && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
  && wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \ 
  && apk add --update ca-certificates \
  && update-ca-certificates \
  && rm -rf /var/cache/apk/*

COPY --from=server-build /github.com/traPtitech/knoq/knoq ./

ENTRYPOINT ./knoq
