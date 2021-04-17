package slack

import (
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/wolf1996/HandWitch/pkg/core"

	log "github.com/sirupsen/logrus"
)

type Bot struct {
	client *socketmode.Client
	api    *slack.Client
	log.Logger
}

func (bot *Bot) buildUrlSelectForm(evt *socketmode.Event) error {

}

func (bot *Bot) socketmodeLiten(logger *log.Logger) {
	go func() {
		for evt := range bot.client.Events {
			log.Debugf("event got! %v", evt)
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Debugf("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				log.Debugf("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				log.Debugf("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				// err := processEventsAPI(client, api, &evt)
				// if err != nil {
				// 	continue
				// }
			case socketmode.EventTypeInteractive:
				// err := processTypeInteractive(client, api, &evt)
				// if err != nil {
				// 	continue
				// }
			case socketmode.EventTypeSlashCommand:
				// err := processSlashCmds(client, api, &evt)
				// if err != nil {
				// 	continue
				// }
			default:
				log.Debugf("Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()
	bot.client.Run()
}

func (bot *Bot) Listen(logger *log.Logger) {
	bot.socketmodeLiten(logger)
}

func NewBot(client *http.Client, appToken string, socketToken string, app core.URLProcessor, logger *log.Logger) (*Bot, error) {
	isDebug := logger.GetLevel() & log.DebugLevel
	api := slack.New(
		appToken,
		slack.OptionDebug(isDebug),
		slack.OptionLog(logger),
		slack.OptionAppLevelToken(appToken),
		slack.OptionHttpClient(client),
	)
	socketClient := socketmode.New(
		api,
		socketmode.OptionDebug(isDebug),
		socketmode.OptionLog(logger),
	)
	return &Bot{
		api:    api,
		client: socketClient,
	}, nil
}
