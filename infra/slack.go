package infra

import (
	"log"
	"os"

	"github.com/nlopes/slack"
)

func NewSlackRTM() *slack.RTM {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)
	logger := log.New(os.Stdout, "[slack] ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	return rtm
}
