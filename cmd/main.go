package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yuki-eto/5ch-slack-bot/config"
	"github.com/yuki-eto/5ch-slack-bot/service"
)

var (
	threadService = service.NewThreadService()
	slackService = service.NewSlackService()
)

func main() {
	log.Println("5ch-bot service started!")
	cfg := config.GetEnvConfig()
	log.Printf("config: %+v\n", cfg)

	threadService = service.NewThreadService()
	slackService = service.NewSlackService()
	slackService.Initialize()

	for {
		log.Println("reload borad threads...")
		mainLoop()
		log.Println("wait...")
		time.Sleep(3 * time.Minute)
	}
}

func mainLoop() {
	threads, err := threadService.GetThreads()
	if err != nil {
		panic(err)
	}

	for _, t := range threads.List {
		newArticleCount := len(t.Articles)
		if newArticleCount == 0 {
			continue
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("%d: %s (%d)", t.ThreadID, t.Title, t.LastReadArticleID))
		text := strings.Join(lines, "\n")
		log.Println(text)
		slackService.SendMessage(text)
		for _, a := range t.Articles {
			text := a.FormatString()
			log.Println(text)
			slackService.SendMessage(text)
			time.Sleep(1 * time.Second)
		}
	}
}
