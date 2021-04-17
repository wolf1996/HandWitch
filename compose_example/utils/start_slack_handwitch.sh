#!/bin/bash

CONFIG_NAME="config.json"

build_config () {
    if [[ "$HANDWITCH_USE_WEBHOOK" ]]; then 
        echo "Attemping to build and use webhook configuration"
        HOOK_CFG=$(jq -n \
            --arg cert "/crt/$SSL_PUB_NAME"\
            --arg key "/crt/$SSL_PRIV_NAME"\
            --arg webhook $HANDWITCH_USE_WEBHOOK \
            -f config_hook_part.json)
        cat config_template.json | jq --argjson HOOK_CFG "$HOOK_CFG" '(. + {hook:$HOOK_CFG})' > $CONFIG_NAME
    else 
        echo "Copy plain polling config"
        cp config_template.json $CONFIG_NAME
    fi
    echo "Final config is: $(cat $CONFIG_NAME)"
}

start_bot () {
    /HandWitch/HandWitch slack --token=$HANDWITCH_TELEGRAMM_TOKEN --config=$CONFIG_NAME
}

build_config
echo "Starting bot"
start_bot