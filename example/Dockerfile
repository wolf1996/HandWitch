FROM golang as build

ENV GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=auto \
    CGO_ENABLED=0

WORKDIR /build/

COPY server.go .

RUN go build -o Server

FROM alpine:3.12 as bin

WORKDIR /server/

COPY --from=build /build/Server .

CMD ["./Server"]
