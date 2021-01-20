FROM golang:1.15-alpine as builder

RUN apk --update add git build-base make

ARG PLATFORM=linux
ARG ARCHITECTURE=amd64
ARG FILES=./cmd

WORKDIR /opt

COPY . .

RUN go mod download && make build

FROM gcr.io/distroless/static

LABEL maintainer="mmadraimov@gmail.com"

ARG DBURL
ENV DBURL $DBURL
WORKDIR /opt
COPY --from=builder /opt/bin/event-timing-server .
COPY --from=builder /opt/docs ./docs

CMD ["./event-timing-server"]
