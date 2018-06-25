package entity

import (
	"sort"
	"time"

	"github.com/yuki-eto/pot-collector"
)

type Thread struct {
	ID                uint64     `db:"id, primarykey, autoincrement"`
	ThreadID          uint64     `db:"thread_id"`
	Title             string     `db:"title, size:255"`
	LastReadArticleID uint32     `db:"last_read_article_id"`
	IsFinished        bool       `db:"is_finished"`
	UpdatedAt         *time.Time `db:"updated_at"`
	CreatedAt         *time.Time `db:"created_at"`
	Articles          []*Article `db:"-"`
}

func NewThread(t *potCollector.Thread) *Thread {
	return &Thread{
		ThreadID:          t.ID,
		Title:             t.Title,
		LastReadArticleID: 1,
		IsFinished:        t.LastArticleID >= 1000,
		Articles:          []*Article{},
	}
}

func (p *Thread) SetArticles(articles *Articles) {
	// 書き込み順に並ぶように slice をソートしてからセットする
	var ids []uint32
	for id := range articles.List {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	for _, id := range ids {
		p.Articles = append(p.Articles, articles.List[id])
	}
}

func (p *Thread) ReplaceByCurrentThread(pt *potCollector.Thread) {
	p.LastReadArticleID = pt.LastArticleID
	p.IsFinished = pt.LastArticleID >= 1000
}

type Threads struct {
	List  map[uint64]*Thread
	Count int
}

func NewThreads() *Threads {
	return &Threads{
		List:  map[uint64]*Thread{},
		Count: 0,
	}
}

func (p *Threads) Append(t *Thread) {
	_, exists := p.List[t.ThreadID]
	p.List[t.ThreadID] = t
	if !exists {
		p.Count++
	}
}

func (p *Threads) GetIDs() []uint64 {
	var keys []uint64
	for k := range p.List {
		keys = append(keys, k)
	}
	return keys
}

func (p *Threads) Get(id uint64) *Thread {
	return p.List[id]
}
