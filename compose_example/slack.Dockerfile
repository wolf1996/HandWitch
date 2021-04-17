FROM glazedcurd/handwitch

RUN apk update \
 && apk add jq \
 && rm -rf /var/cache/apk/*

WORKDIR /HandWitch/

COPY ./utils utils
COPY ./config_hook_part.json .
COPY ./config_template.json .

EXPOSE 8443

CMD ["sh", "./utils/start_slack_handwitch.sh"]