package infra

import (
	"log"
	"os"

	"github.com/nlopes/slack"
	"strings"
	"github.com/yuki-eto/5ch-slack-bot/config"
)

type Slack struct {
	rtm *slack.RTM
	ReceivedEvents chan *MessageEvent
}

type MessageEvent struct {
	OriginalText string
	Messages []string
}

func NewSlack() *Slack {
	cfg := config.GetEnvConfig()
	api := slack.New(cfg.SlackToken)
	logger := log.New(os.Stdout, "[slack] ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	return &Slack{
		rtm: rtm,
		ReceivedEvents: make(chan *MessageEvent, 5),
	}
}

func (p *Slack) HandleAction() {
	for msg := range p.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			log.Println("Invalid credentials")
		case *slack.MessageEvent:
			event := &MessageEvent{
				OriginalText: ev.Msg.Text,
				Messages: strings.Split(strings.TrimSpace(ev.Msg.Text), " "),
			}
			p.ReceivedEvents <- event
		}
	}
}

func (p *Slack) SendMessage(text string) {
	cfg := config.GetEnvConfig()
	msg := p.rtm.NewOutgoingMessage(text, cfg.SlackChannel)
	p.rtm.SendMessage(msg)
}
