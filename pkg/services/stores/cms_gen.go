// This file is generated - Do Not Edit.

package stores

import (
	"context"
	"fmt"
	comm "hyyl.xyz/cupola/andvari/models/comm"
	"hyyl.xyz/cupola/scaffold/pkg/models/cms1"
)

// type Article = cms1.Article
// type ArticleBasic = cms1.ArticleBasic
// type ArticleSet = cms1.ArticleSet
// type Articles = cms1.Articles
// type Clause = cms1.Clause
// type ClauseBasic = cms1.ClauseBasic
// type ClauseSet = cms1.ClauseSet
// type Clauses = cms1.Clauses

func init() {
	alltables = append(alltables, &cms1.Article{}, &cms1.Clause{})
}

type ContentStore interface {
	ListClause(ctx context.Context, spec *ClauseSpec) (data cms1.Clauses, total int, err error)
	GetClause(ctx context.Context, id string) (obj *cms1.Clause, err error)
	PutClause(ctx context.Context, id string, in cms1.ClauseSet) (nid string, err error)
	DeleteClause(ctx context.Context, id string) error

	ListArticle(ctx context.Context, spec *ArticleSpec) (data cms1.Articles, total int, err error)
	GetArticle(ctx context.Context, id string) (obj *cms1.Article, err error)
	CreateArticle(ctx context.Context, in cms1.ArticleBasic) (obj *cms1.Article, err error)
	UpdateArticle(ctx context.Context, id string, in cms1.ArticleSet) error
	DeleteArticle(ctx context.Context, id string) error
}

type ClauseSpec struct {
	comm.PageSpec
	MDftSpec
}
type ArticleSpec struct {
	comm.PageSpec
	MDftSpec

	Author string `extensions:"x-order=A" form:"author" json:"author"` // 作者
}

func (spec *ArticleSpec) Sift(q *ormQuery) (*ormQuery, error) {
	q, _ = spec.MDftSpec.Sift(q)
	q, _ = siftEquel(q, "author", spec.Author, false)

	return q, nil
}

type contentStore struct {
	w *Wrap
}

func (s *contentStore) ListClause(ctx context.Context, spec *ClauseSpec) (data cms1.Clauses, total int, err error) {
	total, err = queryPager(spec, s.w.db.Model(&data).Apply(spec.Sift))
	return
}
func (s *contentStore) GetClause(ctx context.Context, id string) (obj *cms1.Clause, err error) {
	obj = new(cms1.Clause)
	err = getModelWithPKID(ctx, s.w.db, obj, id)
	return
}
func (s *contentStore) PutClause(ctx context.Context, id string, in cms1.ClauseSet) (nid string, err error) {
	obj := new(cms1.Clause)
	_ = obj.SetID(id)
	obj.SetWith(in)
	err = dbStoreSimple(ctx, s.w.db, obj)
	nid = obj.StringID()
	return
}
func (s *contentStore) DeleteClause(ctx context.Context, id string) error {
	obj := new(cms1.Clause)
	if !obj.SetID(id) {
		return fmt.Errorf("id: '%s' is invalid", id)
	}
	return s.w.db.OpDeleteAny(ctx, "cms_clause", obj.ID)
}

func (s *contentStore) ListArticle(ctx context.Context, spec *ArticleSpec) (data cms1.Articles, total int, err error) {
	q := s.w.db.Model(&data).Apply(spec.Sift)
	tss := s.w.db.GetTsSpec()
	tss.SetFallback("title", "content")
	total, err = queryPager(spec, q.Apply(tss.Sift))
	return
}
func (s *contentStore) GetArticle(ctx context.Context, id string) (obj *cms1.Article, err error) {
	obj = new(cms1.Article)
	err = getModelWithPKID(ctx, s.w.db, obj, id)
	return
}
func (s *contentStore) CreateArticle(ctx context.Context, in cms1.ArticleBasic) (obj *cms1.Article, err error) {
	obj = &cms1.Article{
		ArticleBasic: in,
	}
	s.w.opModelMeta(ctx, obj, obj.MetaDiff)
	if tscfg, ok := s.w.db.GetTsCfg(); ok {
		obj.TsCfgName = tscfg
	}
	err = s.w.db.RunInTransaction(ctx, func(tx *pgTx) (err error) {
		if err = dbBeforeSaveArticle(ctx, tx, obj); err != nil {
			return err
		}
		err = dbInsert(ctx, tx, obj)
		return err
	})
	return
}
func (s *contentStore) UpdateArticle(ctx context.Context, id string, in cms1.ArticleSet) error {
	exist := new(cms1.Article)
	if err := getModelWithPKID(ctx, s.w.db, exist, id); err != nil {
		return err
	}
	_ = exist.SetWith(in)
	return s.w.db.RunInTransaction(ctx, func(tx *pgTx) (err error) {
		if err = dbBeforeSaveArticle(ctx, tx, exist); err != nil {
			return
		}
		return dbUpdate(ctx, tx, exist)
	})
}
func (s *contentStore) DeleteArticle(ctx context.Context, id string) error {
	obj := new(cms1.Article)
	if err := getModelWithPKID(ctx, s.w.db, obj, id); err != nil {
		return err
	}
	return s.w.db.RunInTransaction(ctx, func(tx *pgTx) (err error) {
		err = dbDeleteT(ctx, tx, s.w.db.Schema(), s.w.db.SchemaCrap(), "cms_article", obj.ID)
		if err != nil {
			return
		}
		return dbAfterDeleteArticle(ctx, tx, obj)
	})
}
