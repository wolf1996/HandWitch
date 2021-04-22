FROM golang as build

ENV GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=auto \
    CGO_ENABLED=0

WORKDIR /build/

COPY . .

RUN go build

FROM alpine:3.12 as bin

RUN apk update \
 && apk add jq \
 && rm -rf /var/cache/apk/*

WORKDIR /HandWitch/

COPY --from=build /build/HandWitch .
