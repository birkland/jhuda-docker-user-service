FROM golang:1.13-alpine AS builder

RUN apk update && apk add --no-cache git

ENV  GO111MODULE=on \
     CGO_ENABLED=0 

WORKDIR /root
COPY . .
RUN go build -trimpath

FROM alpine:3.10
COPY --from=builder /root/jhuda-docker-user-service /
COPY docker/entrypoint.sh /

RUN chmod 700 /entrypoint.sh

CMD /entrypoint.sh