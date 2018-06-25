package entity

import (
	"encoding/json"
	"time"

	"github.com/yuki-eto/pot-collector"
)

type Article struct {
	ID                  uint64     `db:"id, primarykey, autoincrement"`
	ThreadID            uint64     `db:"thread_id"`
	ArticleID           uint32     `db:"article_id"`
	Name                string     `db:"name, size:255"`
	WroteAt             *time.Time `db:"wrote_at"`
	UID                 string     `db:"uid, size:255"`
	Text                string     `db:"text, size:10000"`
	AnchorArticleIDsStr string     `db:"anchor_article_ids"`
	UpdatedAt           *time.Time `db:"updated_at"`
	CreatedAt           *time.Time `db:"created_at"`
}

func NewArticle(threadID uint64, a *potCollector.Article) *Article {
	newArticle := &Article{
		ThreadID:  threadID,
		ArticleID: a.ID,
		Name:      a.Name,
		WroteAt:   a.Date,
		UID:       a.UID,
		Text:      a.Text,
	}
	newArticle.SetAnchorArticleIDs(a.AnchorArticleIDs)
	return newArticle
}

func (p *Article) GetAnchorArticleIDs() ([]uint32, error) {
	var ids []uint32
	err := json.Unmarshal([]byte(p.AnchorArticleIDsStr), &ids)
	return ids, err
}

func (p *Article) SetAnchorArticleIDs(ids []uint32) error {
	bytes, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	p.AnchorArticleIDsStr = string(bytes)
	return nil
}

type Articles struct {
	List  map[uint32]*Article
	Count int
}

func NewArticles() *Articles {
	return &Articles{
		List:  map[uint32]*Article{},
		Count: 0,
	}
}

func (p *Articles) Append(t *Article) {
	p.List[t.ArticleID] = t
	p.Count++
}

func (p *Articles) GetIDs() []uint32 {
	var keys []uint32
	for k := range p.List {
		keys = append(keys, k)
	}
	return keys
}

func (p *Articles) Get(id uint32) *Article {
	return p.List[id]
}
