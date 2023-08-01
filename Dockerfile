# syntax=docker/dockerfile:1

# for production

FROM golang:1-alpine as server-build

WORKDIR /github.com/traPtitech/knoq

COPY go.mod go.sum ./
ENV GO111MODULE=on
RUN go mod download
COPY ./ ./

RUN go build -o knoq

FROM alpine:latest

WORKDIR /app

RUN apk --update add tzdata \
  && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
  && apk add --update ca-certificates \
  && update-ca-certificates \
  && rm -rf /var/cache/apk/*

COPY --from=server-build /github.com/traPtitech/knoq/knoq ./

ARG knoq_version=dev
ARG knoq_revision=local
ENV KNOQ_VERSION=${knoq_version}
ENV KNOQ_REVISION=${knoq_revision}

ENTRYPOINT ./knoq
