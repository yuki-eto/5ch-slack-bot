package dao

import (
	"time"

	"gopkg.in/Masterminds/squirrel.v1"
	"gopkg.in/gorp.v2"

	"github.com/yuki-eto/5ch-slack-bot/entity"
	"github.com/yuki-eto/5ch-slack-bot/infra"
)

type Thread interface {
	getTableName() string
	getDbMap() *gorp.DbMap
	SelectList([]uint64) (*entity.Threads, error)
	Insert(*entity.Thread) (*entity.Thread, error)
	Update(*entity.Thread) (*entity.Thread, error)
}

type ThreadImpl struct {
	tableName string
	dbMap     *gorp.DbMap
}

func (p *ThreadImpl) getTableName() string {
	return p.tableName
}
func (p *ThreadImpl) getDbMap() *gorp.DbMap {
	return p.dbMap
}

func init() {
	d := NewThread()
	dbMap := d.getDbMap()
	table := dbMap.AddTableWithName(entity.Thread{}, d.getTableName())
	table.SetKeys(true, "ID")
	table.AddIndex("unique_thread", "", []string{"thread_id"}).SetUnique(true)
}

func NewThread() Thread {
	return &ThreadImpl{
		tableName: "threads",
		dbMap:     infra.NewDBMap("default"),
	}
}

func (p *ThreadImpl) SelectList(id []uint64) (*entity.Threads, error) {
	sql, args, err := squirrel.Select("*").
		From(p.tableName).
		Where(squirrel.Eq{"thread_id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var thread entity.Thread
	rows, err := p.dbMap.Select(&thread, sql, args...)
	if err != nil {
		return nil, err
	}

	threads := entity.NewThreads()
	for _, row := range rows {
		switch f := row.(type) {
		case *entity.Thread:
			threads.Append(f)
		}
	}
	return threads, nil
}

func (p *ThreadImpl) Insert(t *entity.Thread) (*entity.Thread, error) {
	now := time.Now()
	t.UpdatedAt = &now
	t.CreatedAt = &now
	err := p.dbMap.Insert(t)
	return t, err
}

func (p *ThreadImpl) Update(t *entity.Thread) (*entity.Thread, error) {
	now := time.Now()
	t.UpdatedAt = &now
	_, err := p.dbMap.Update(t)
	return t, err
}
