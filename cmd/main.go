package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/yuki-eto/5ch-slack-bot/dao"
	"github.com/yuki-eto/5ch-slack-bot/entity"
	"github.com/yuki-eto/5ch-slack-bot/infra"
	"github.com/yuki-eto/pot-collector"
)

type Slack struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

var (
	threadListURL = os.Getenv("THREAD_LIST_URL")
	threadBaseURL = os.Getenv("THREAD_BASE_URL")
	rtm           *slack.RTM
)

func main() {
	log.Println("daemon started!")
	rtm = infra.NewSlackRTM()
	go handlingRTM(rtm)

	for {
		log.Println("reload borad threads...")
		mainLoop()
		log.Println("wait...")
		time.Sleep(3 * time.Minute)
	}
}

func mainLoop() {
	threads := loadThreadArticles()

	for _, t := range threads.List {
		newArticleCount := len(t.Articles)
		if newArticleCount == 0 {
			continue
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("%s (%d)", t.Title, t.LastReadArticleID))
		lines = append(lines, fmt.Sprintf("%s%d/", threadBaseURL, t.ThreadID))
		sendSlack(strings.Join(lines, "\n"))
		for _, a := range t.Articles {
			var lines []string
			lines = append(lines, "```")
			lines = append(lines, fmt.Sprintf("%d: %s %v UID:%s", a.ArticleID, a.Name, a.WroteAt, a.UID))
			lines = append(lines, a.Text)
			lines = append(lines, "```")

			sendSlack(strings.Join(lines, "\n"))
			time.Sleep(1 * time.Second)
		}
	}
}

func sendSlack(text string) {
	channel := os.Getenv("SLACK_CHANNEL")
	msg := rtm.NewOutgoingMessage(text, channel)
	rtm.SendMessage(msg)
}

func loadThreadArticles() *entity.Threads {
	board := potCollector.NewBoard()
	log.Println("loading thread list...")
	doc, err := board.LoadThreadListDocument(threadListURL)
	if err != nil {
		panic(err)
	}
	if err := board.LoadThreads(doc); err != nil {
		panic(err)
	}

	board.Threads.FilterThread(func(t *potCollector.Thread) bool {
		return strings.Contains(t.Title, os.Getenv("THREAD_NAME_CONTAINS"))
	})

	threadDao := dao.NewThread()
	articleDao := dao.NewArticle()
	newThreads := entity.NewThreads()
	for _, thread := range board.Threads.List {
		newThread := entity.NewThread(thread)
		newThreads.Append(newThread)
	}

	log.Println("loading old thread statuses...")
	oldThreads, err := threadDao.SelectList(newThreads.GetIDs())
	if err != nil {
		panic(err)
	}

	for _, thread := range board.Threads.List {
		log.Printf("processing thread %d: %s\n", thread.ID, thread.Title)

		id := thread.ID
		oldThread, hasOldThread := oldThreads.List[id]
		if hasOldThread {
			newThreads.Append(oldThread)
			if thread.LastArticleID == oldThread.LastReadArticleID {
				log.Println("no updated")
				continue
			}
		}

		if hasOldThread {
			thread.LastReadArticleID = oldThread.LastReadArticleID
		}

		log.Printf("loading thread articles : %d\n", thread.ID)
		doc, err := thread.LoadArticleDocument(threadBaseURL)
		if err != nil {
			panic(err)
		}
		if err := thread.LoadArticles(doc); err != nil {
			panic(err)
		}

		log.Println("store thread and articles...")
		targetThread := newThreads.Get(thread.ID)
		newArticles := entity.NewArticles()
		for _, article := range thread.Articles.List {
			if hasOldThread && article.ID <= oldThread.LastReadArticleID {
				continue
			}
			if article.IsOver1000 {
				continue
			}
			newArticle := entity.NewArticle(thread.ID, article)
			if _, err := articleDao.Insert(newArticle); err != nil {
				log.Fatalf("%+v\n", err)
			}
			newArticles.Append(newArticle)
		}
		targetThread.SetArticles(newArticles)

		if hasOldThread {
			oldThread.ReplaceByCurrentThread(thread)
			if _, err := threadDao.Update(oldThread); err != nil {
				panic(err)
			}
			newThreads.Append(oldThread)
		} else {
			newThread := newThreads.Get(id)
			newThread.ReplaceByCurrentThread(thread)
			if _, err := threadDao.Insert(newThread); err != nil {
				panic(err)
			}
		}

		log.Printf("end : %d\n", thread.ID)
	}

	return newThreads
}

func handlingRTM(rtm *slack.RTM) {
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.LatencyReport:
			fmt.Printf("Current Latency: %v\n", ev.Value)
		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			fmt.Println("Invalid credentials")
		}
	}
}
