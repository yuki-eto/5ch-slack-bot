package service

import (
	"log"
	"strings"

	"github.com/yuki-eto/pot-collector"
	"github.com/yuki-eto/5ch-slack-bot/entity"
	"github.com/yuki-eto/5ch-slack-bot/dao"
	"github.com/yuki-eto/5ch-slack-bot/config"
)

type ThreadService interface {
	GetThreads() (*entity.Threads, error)
	processThread(*potCollector.Thread, *entity.Threads, *entity.Threads) error
}

type ThreadServiceImpl struct {
	threadDao dao.Thread
	articleDao dao.Article
}

func NewThreadService() ThreadService {
	return &ThreadServiceImpl{
		threadDao: dao.NewThread(),
		articleDao: dao.NewArticle(),
	}
}

func (p *ThreadServiceImpl) GetThreads() (*entity.Threads, error) {
	board := potCollector.NewBoard()
	log.Println("loading thread list...")
	cfg := config.GetEnvConfig()
	doc, err := board.LoadThreadListDocument(cfg.ThreadListURL)
	if err != nil {
		return nil, err
	}
	if err := board.LoadThreads(doc); err != nil {
		return nil, err
	}

	board.Threads.FilterThread(func(t *potCollector.Thread) bool {
		return strings.Contains(t.Title, cfg.ThreadNameContains)
	})

	newThreads := entity.NewThreads()
	for _, thread := range board.Threads.List {
		newThread := entity.NewThread(thread)
		newThreads.Append(newThread)
	}

	log.Println("loading old thread statuses...")
	oldThreads, err := p.threadDao.SelectList(newThreads.GetIDs())
	if err != nil {
		return nil, err
	}

	for _, thread := range board.Threads.List {
		if err := p.processThread(thread, newThreads, oldThreads); err != nil {
			return nil, err
		}
	}

	return newThreads, nil
}

func (p* ThreadServiceImpl) processThread(thread *potCollector.Thread, newThreads *entity.Threads, oldThreads *entity.Threads) error {
	log.Printf("processing thread %d: %s\n", thread.ID, thread.Title)

	id := thread.ID
	oldThread, hasOldThread := oldThreads.List[id]
	if hasOldThread {
		newThreads.Append(oldThread)
		if thread.LastArticleID == oldThread.LastReadArticleID {
			log.Println("no updated")
			return nil
		}
	}

	if hasOldThread {
		thread.LastReadArticleID = oldThread.LastReadArticleID
	}

	log.Printf("loading thread articles : %d\n", thread.ID)
	cfg := config.GetEnvConfig()
	doc, err := thread.LoadArticleDocument(cfg.ThreadBaseURL)
	if err != nil {
		return err
	}
	if err := thread.LoadArticles(doc); err != nil {
		return err
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
		if _, err := p.articleDao.Insert(newArticle); err != nil {
			log.Printf("%+v\n", err)
			if !strings.Contains(err.Error(), "UNIQUE") {
				return err
			}
		}
		newArticles.Append(newArticle)
	}
	targetThread.SetArticles(newArticles)

	if hasOldThread {
		oldThread.ReplaceByCurrentThread(thread)
		newThreads.Append(oldThread)
		if _, err := p.threadDao.Update(oldThread); err != nil {
			return err
		}
	} else {
		newThread := newThreads.Get(id)
		newThread.ReplaceByCurrentThread(thread)
		if _, err := p.threadDao.Insert(newThread); err != nil {
			return err
		}
	}

	log.Printf("end : %d\n", thread.ID)
	return nil
}
