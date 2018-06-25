package dao

import (
	"time"

	"gopkg.in/Masterminds/squirrel.v1"
	"gopkg.in/gorp.v2"

	"github.com/yuki-eto/5ch-slack-bot/entity"
	"github.com/yuki-eto/5ch-slack-bot/infra"
)

type Article interface {
	getTableName() string
	getDbMap() *gorp.DbMap
	Select(uint64, uint32) (*entity.Article, error)
	Insert(*entity.Article) (*entity.Article, error)
}

type ArticleImpl struct {
	tableName string
	dbMap     *gorp.DbMap
}

func (p *ArticleImpl) getTableName() string {
	return p.tableName
}
func (p *ArticleImpl) getDbMap() *gorp.DbMap {
	return p.dbMap
}

func init() {
	d := NewArticle()
	dbMap := d.getDbMap()
	table := dbMap.AddTableWithName(entity.Article{}, d.getTableName())
	table.SetKeys(true, "ID")
	table.AddIndex("unique_article", "", []string{"thread_id", "article_id"}).SetUnique(true)
}

func NewArticle() Article {
	return &ArticleImpl{
		tableName: "articles",
		dbMap:     infra.NewDBMap("default"),
	}
}

func (p *ArticleImpl) Select(threadId uint64, articleId uint32) (*entity.Article, error) {
	sql, args, err := squirrel.Select("*").
		From(p.tableName).
		Where(squirrel.Eq{"thread_id": threadId, "article_id": articleId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var article entity.Article
	if err := p.dbMap.SelectOne(&article, sql, args...); err != nil {
		return nil, err
	}
	return &article, nil
}

func (p *ArticleImpl) Insert(t *entity.Article) (*entity.Article, error) {
	now := time.Now()
	t.UpdatedAt = &now
	t.CreatedAt = &now
	err := p.dbMap.Insert(t)
	return t, err
}
