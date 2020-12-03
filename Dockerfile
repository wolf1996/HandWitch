FROM golang as build

ENV GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=auto \
    CGO_ENABLED=0

WORKDIR /build/

COPY . .

RUN go build

FROM alpine:latest as bin

WORKDIR /HandWitch/

COPY --from=build /build/HandWitch .
COPY example/config.json .
COPY example/whitelist.json .
COPY example/descriptions.yaml .
COPY compose_example/utils utils
COPY compose_example/config_hook_part.json .
COPY compose_example/config_template.json .

EXPOSE 8443

CMD ["sh", "./utils/start_handwitch.sh"]
