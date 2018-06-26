package service

import (
	"fmt"
	"strconv"

	"github.com/yuki-eto/5ch-slack-bot/infra"
	"github.com/yuki-eto/5ch-slack-bot/dao"
	"github.com/yuki-eto/5ch-slack-bot/entity"
	"github.com/yuki-eto/5ch-slack-bot/config"
)

type SlackService interface {
	Initialize()
	SendMessage(string)
	handleMessageAction(*infra.MessageEvent)
}

type SlackServiceImpl struct {
	slack *infra.Slack
	articleDao dao.Article
}

func NewSlackService() SlackService {
	return &SlackServiceImpl{
		slack: infra.NewSlack(),
		articleDao: dao.NewArticle(),
	}
}

func (p *SlackServiceImpl) Initialize() {
	go p.slack.HandleAction()
	go func () {
		for event := range p.slack.ReceivedEvents {
			p.handleMessageAction(event)
		}
	}()
}

func (p *SlackServiceImpl) SendMessage(text string) {
	p.slack.SendMessage(text)
}

func (p *SlackServiceImpl) handleMessageAction(ev *infra.MessageEvent) {
	messages := ev.Messages
	if len(messages) < 2 {
		return
	}
	cfg := config.GetEnvConfig()
	toBotString := fmt.Sprintf("<@%s>", cfg.SlackBotID)
	if messages[0] != toBotString {
		return
	}

	action := messages[1]
	switch action {
	case "article":
		if len(messages) < 4 {
			p.slack.SendMessage("[error] not enough arguments for getting article")
			return
		}
		args := messages[2:]
		article, err := p.getArticle(args[0], args[1])
		if err != nil {
			p.slack.SendMessage(fmt.Sprintf("[error] %v", err))
			return
		}
		p.slack.SendMessage(article.FormatString())
	}
}

func (p *SlackServiceImpl) getArticle(str1, str2 string) (*entity.Article, error) {
	threadID, err := strconv.ParseUint(str1, 10, 64)
	if err != nil {
		return nil, err
	}
	articleID, err := strconv.ParseUint(str2, 10, 32)
	if err != nil {
		return nil, err
	}
	return p.articleDao.Select(threadID, uint32(articleID))
}


