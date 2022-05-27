// This file is generated - Do Not Edit.

package stores

import (
	"context"
	comm "hyyl.xyz/cupola/aurora/pkg/models/comm"
	"hyyl.xyz/cupola/scaffold/pkg/models/cms1"
)

type Article = cms1.Article
type ArticleSet = cms1.ArticleSet
type Articles = cms1.Articles
type Clause = cms1.Clause
type ClauseSet = cms1.ClauseSet
type Clauses = cms1.Clauses

type ContantStore interface {
	ListArticle(ctx context.Context, spec *ArticleSpec) (data []cms1.Article, total int, err error)
	GetArticle(ctx context.Context, id string) (obj *cms1.Article, err error)
	CreateArticle(ctx context.Context, in *cms1.ArticleBasic) (obj *cms1.Article, err error)
	UpdateArticle(ctx context.Context, id string, in *cms1.ArticleSet) (err error)
	DeleteArticle(ctx context.Context, id string) error

	ListClause(ctx context.Context, spec *ClauseSpec) (data []cms1.Clause, total int, err error)
	GetClause(ctx context.Context, id string) (obj *cms1.Clause, err error)
	PutClause(ctx context.Context, id string, in *cms1.ClauseSet) (err error)
	DeleteClause(ctx context.Context, id string) error
}

type ArticleSpec struct {
	comm.PageSpec
	MDftSpec

	Title string `form:"title" json:"title"` // 标题
}

func (spec *ArticleSpec) Sift(q *ormQuery) (*ormQuery, error) {
	q, _ = spec.MDftSpec.Sift(q)
	q, _ = siftEquel(q, "title", spec.Title, false)

	return q, nil
}

type ClauseSpec struct {
	comm.PageSpec
	MDftSpec
}

type contentStore struct {
	w *Wrap
}

func (s *contentStore) ListArticle(ctx context.Context, spec *ArticleSpec) (data []cms1.Article, total int, err error) {
	total, err = queryPager(spec, s.w.db.Model(&data).Apply(spec.Sift))
	return
}
func (s *contentStore) GetArticle(ctx context.Context, id string) (obj *cms1.Article, err error) {
	obj = new(cms1.Article)
	err = getModelWithPKOID(s.w.db, obj, id)
	return
}
func (s *contentStore) CreateArticle(ctx context.Context, in *cms1.ArticleBasic) (obj *cms1.Article, err error) {
	obj = &cms1.Article{
		ArticleBasic: *in,
	}
	err = dbInsert(ctx, s.w.db, obj)
	return
}
func (s *contentStore) UpdateArticle(ctx context.Context, id string, in *cms1.ArticleSet) (err error) {
	exist := new(cms1.Article)
	err = getModelWithPKOID(s.w.db, exist, id)
	if err != nil {
		return
	}
	cs := exist.SetWith(in)
	if len(cs) == 0 {
		return
	}
	err = dbUpdate(ctx, s.w.db, exist, cs...)
	return
}
func (s *contentStore) DeleteArticle(ctx context.Context, id string) error {
	return s.w.db.OpDelete(ctx, "cms_article", id)
}

func (s *contentStore) ListClause(ctx context.Context, spec *ClauseSpec) (data []cms1.Clause, total int, err error) {
	total, err = queryPager(spec, s.w.db.Model(&data).Apply(spec.Sift))
	return
}
func (s *contentStore) GetClause(ctx context.Context, id string) (obj *cms1.Clause, err error) {
	obj = new(cms1.Clause)
	err = getModelWithPKOID(s.w.db, obj, id)
	return
}
func (s *contentStore) PutClause(ctx context.Context, id string, in *cms1.ClauseSet) (err error) {
	obj := new(cms1.Clause)
	obj.SetID(id)
	cs := obj.SetWith(in)
	err = dbStoreSimple(ctx, s.w.db, obj, cs...)
	return
}
func (s *contentStore) DeleteClause(ctx context.Context, id string) error {
	return s.w.db.OpDelete(ctx, "cms_clause", id)
}
