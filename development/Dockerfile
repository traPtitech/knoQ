FROM golang:1.15.6-alpine
WORKDIR /srv/knoq

ENV DOCKERIZE_VERSION v0.6.1
RUN apk add --no-cache openssl \
 && wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
 && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
 && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

RUN apk add --update --no-cache git \
  &&  go get -u github.com/pilu/fresh

COPY go.mod ./
COPY go.sum ./
ENV GO111MODULE=on
RUN go mod download
COPY ./ ./
ENTRYPOINT fresh -c development/fresh.conf
