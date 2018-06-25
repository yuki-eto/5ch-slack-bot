package infra

import (
	"log"
	"os"

	"github.com/nlopes/slack"
)

type Slack struct {
	rtm *slack.RTM
	Actions chan *slack.MessageEvent
}

func NewSlack() *Slack {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)
	logger := log.New(os.Stdout, "[slack] ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	return &Slack{
		rtm: rtm,
		Actions: make(chan *slack.MessageEvent, 5),
	}
}

func (p *Slack) HandleAction() {
	rtm := p.rtm
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.LatencyReport:
			log.Printf("Current Latency: %v\n", ev.Value)
		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			log.Println("Invalid credentials")
		case *slack.MessageEvent:
			p.Actions <- ev
		}
	}
}

func (p *Slack) SendMessage(text string) {
	channel := os.Getenv("SLACK_CHANNEL")
	msg := p.rtm.NewOutgoingMessage(text, channel)
	p.rtm.SendMessage(msg)
}
