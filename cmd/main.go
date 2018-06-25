package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	slackLib "github.com/nlopes/slack"
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
	slack           *infra.Slack
	threadDao       dao.Thread
	articleDao      dao.Article
)

func main() {
	log.Println("daemon started!")
	slack = infra.NewSlack()
	go slack.HandleAction()
	go func () {
		for action := range slack.Actions {
			handleMessageAction(action)
		}
	}()

	threadDao = dao.NewThread()
	articleDao = dao.NewArticle()

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
		lines = append(lines, fmt.Sprintf("%d: %s (%d)", t.ThreadID, t.Title, t.LastReadArticleID))
		lines = append(lines, fmt.Sprintf("%s%d/", threadBaseURL, t.ThreadID))
		slack.SendMessage(strings.Join(lines, "\n"))
		for _, a := range t.Articles {
			slack.SendMessage(formatArticle(a))
			time.Sleep(1 * time.Second)
		}
	}
}

func formatArticle(a *entity.Article) string {
	var lines []string
	lines = append(lines, "```")
	lines = append(lines, fmt.Sprintf("%d %d: %s %v UID:%s", a.ThreadID, a.ArticleID, a.Name, a.WroteAt, a.UID))
	lines = append(lines, a.Text)
	lines = append(lines, "```")
	lines = append(lines, fmt.Sprintf("%s%d/%d", threadBaseURL, a.ThreadID, a.ArticleID))
	return strings.Join(lines, "\n")
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
				log.Printf("%+v\n", err)
				if !strings.Contains(err.Error(), "UNIQUE") {
					panic(err)
				}
			}
			newArticles.Append(newArticle)
		}
		targetThread.SetArticles(newArticles)

		if hasOldThread {
			oldThread.ReplaceByCurrentThread(thread)
			newThreads.Append(oldThread)
			if _, err := threadDao.Update(oldThread); err != nil {
				panic(err)
			}
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

func handleMessageAction(ev *slackLib.MessageEvent) {
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", os.Getenv("SLACK_BOT_ID"))) {
		return
	}
	messages := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	if len(messages) == 0 {
		return
	}
	keyword := messages[0]

	switch keyword {
	case "article":
		threadID, err := strconv.ParseUint(messages[1], 10, 64)
		if err != nil {
			slack.SendMessage(fmt.Sprintf("> [error] %v\n", err))
			return
		}
		articleID, err := strconv.ParseUint(messages[2], 10, 32)
		if err != nil {
			slack.SendMessage(fmt.Sprintf("> [error] %v\n", err))
			return
		}
		article, err := articleDao.Select(threadID, uint32(articleID))
		if err != nil {
			slack.SendMessage(fmt.Sprintf("> [error] %v\n", err))
			return
		}
		slack.SendMessage(formatArticle(article))
	}
}
