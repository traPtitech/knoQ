# syntax=docker/dockerfile:1

#
# build stage
#
FROM --platform=$BUILDPLATFORM golang:1-alpine as builder

WORKDIR /app

ENV GO111MODULE=on
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod

COPY go.mod go.sum ./
RUN --mount=type=cache,target=${GOCACHE} \
  --mount=type=cache,target=${GOMODCACHE} \
  go mod download

COPY ./ ./
RUN --mount=type=cache,target=${GOCACHE} \
  --mount=type=cache,target=${GOMODCACHE} \
  go build -o /app/knoq

# static files
RUN mkdir -p /app/web \
  && apk add --no-cache curl \
  && curl -L -Ss https://github.com/traPtitech/knoQ-UI/releases/latest/download/dist.tar.gz \
  | tar zxv -C /app/web
# Google Calendar API needs service.json
RUN touch /app/service.json

#
# runtime stage
#
FROM gcr.io/distroless/static-debian11:latest

WORKDIR /app

# COPY --from=builder /app/knoq /app/web/ /app/service.json /app/
COPY --from=builder /app/knoq /app/
COPY --from=builder /app/web/ /app/web/
COPY --from=builder /app/service.json /app/

ARG knoq_version=dev
ARG knoq_revision=local
ENV KNOQ_VERSION=${knoq_version}
ENV KNOQ_REVISION=${knoq_revision}

CMD ["/app/knoq"]
