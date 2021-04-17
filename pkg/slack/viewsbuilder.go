package slack

import (
	"github.com/slack-go/slack"
	"github.com/wolf1996/HandWitch/pkg/core"

	log "github.com/sirupsen/logrus"
)

const (
	UrlSelectionViewId    = "UrlSelection"
	ChannelSelectBlockId  = "SelectChannel"
	ChannelSelectActionId = "SelectChannelAction"
	URLSelectBlockId      = "SelectBlock"
	URLSelectActionId     = "SelectBlockAction"
)

func buildUrlSelectionBlock(processor *core.URLProcessor, logger *log.Logger) (slack.ModalViewRequest, error) {
	names, err := processor.GetHandsNames()
	if err != nil {
		return slack.ModalViewRequest{}, err
	}

	buildUrlOptinons := func() []*slack.OptionBlockObject {
		result := make([]*slack.OptionBlockObject, len(names))
		for _, urlName := range names {
			result = append(result, &slack.OptionBlockObject{
				Text: &slack.TextBlockObject{
					Type: "plain_text",
					Text: urlName,
				},
				Value: urlName,
			})
		}
		return result
	}
	request := slack.ModalViewRequest{
		Type:       "modal",
		CallbackID: UrlSelectionViewId,
		Title: &slack.TextBlockObject{
			Type: "plain_text",
			Text: "Построение запроса",
		},
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewInputBlock(
					ChannelSelectBlockId,
					&slack.TextBlockObject{
						Type: "plain_text",
						Text: "Куда отправить сообщение:",
					},
					&slack.SelectBlockElement{
						Type:     slack.MultiOptTypeConversations,
						ActionID: ChannelSelectActionId,
						Placeholder: &slack.TextBlockObject{
							Type: slack.PlainTextType,
							Text: "channel to send message",
						},
						DefaultToCurrentConversation: true,
					},
				),
				slack.NewActionBlock(
					URLSelectBlockId,
					&slack.SelectBlockElement{
						Type:     slack.OptTypeStatic,
						ActionID: URLSelectActionId,
						Placeholder: &slack.TextBlockObject{
							Type: slack.PlainTextType,
							Text: "some string options",
						},
						Options: buildUrlOptinons(),
					},
				),
			},
		},
	}
	return request, nil
}
