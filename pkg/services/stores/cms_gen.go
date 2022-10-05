// This file is generated - Do Not Edit.

package stores

import (
	"context"
	"fmt"
	comm "github.com/cupogo/andvari/models/comm"
	utils "github.com/cupogo/andvari/utils"
	"github.com/cupogo/scaffold/pkg/models/cms1"
)

// type Article = cms1.Article
// type ArticleBasic = cms1.ArticleBasic
// type ArticleSet = cms1.ArticleSet
// type Articles = cms1.Articles
// type Attachment = cms1.Attachment
// type AttachmentBasic = cms1.AttachmentBasic
// type AttachmentSet = cms1.AttachmentSet
// type Attachments = cms1.Attachments
// type Clause = cms1.Clause
// type ClauseBasic = cms1.ClauseBasic
// type ClauseSet = cms1.ClauseSet
// type Clauses = cms1.Clauses

func init() {
	RegisterModel((*cms1.Article)(nil), (*cms1.Attachment)(nil), (*cms1.Clause)(nil))
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

	ListAttachment(ctx context.Context, spec *AttachmentSpec) (data cms1.Attachments, total int, err error)
	GetAttachment(ctx context.Context, id string) (obj *cms1.Attachment, err error)
}

type ClauseSpec struct {
	comm.PageSpec
	ModelSpec

	Text string `extensions:"x-order=A" form:"text" json:"text"`
}

func (spec *ClauseSpec) Sift(q *ormQuery) *ormQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = siftMatch(q, "text", spec.Text, false)

	return q
}

type ArticleSpec struct {
	comm.PageSpec
	ModelSpec
	TextSearchSpec

	// 作者
	Author string `extensions:"x-order=A" form:"author" json:"author"`
	// 标题
	Title string `extensions:"x-order=B" form:"title" json:"title"`
	// 新闻时间 + during
	NewsPublish string `extensions:"x-order=C" form:"newsPublish" json:"newsPublish,omitempty"`
	// 状态 (多值逗号分隔)
	Statuses string `extensions:"x-order=D" form:"statuses" json:"statuses,omitempty"`
	// 状态
	Status int16 `extensions:"x-order=E" form:"status" json:"status"`
	// 作者
	AuthorID string `extensions:"x-order=F" form:"authorID" json:"authorID"`
	// 来源 (多值逗号分隔)
	Srcs string `extensions:"x-order=G" form:"srcs" json:"srcs,omitempty"`
	// 来源
	Src string `extensions:"x-order=H" form:"src" json:"src"`
}

func (spec *ArticleSpec) Sift(q *ormQuery) *ormQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = siftICE(q, "author", spec.Author, false)
	q, _ = siftMatch(q, "title", spec.Title, false)
	q, _ = siftDate(q, "news_publish", spec.NewsPublish, true, false)
	if vals, ok := utils.ParseInts(spec.Statuses); ok {
		q = q.Where("status IN(?)", pgIn(vals))
	} else {
		q, _ = siftEquel(q, "status", spec.Status, false)
	}
	q, _ = siftOIDs(q, "author_id", spec.AuthorID, false)
	if vals, ok := utils.ParseStrs(spec.Srcs); ok {
		q = q.Where("src IN(?)", pgIn(vals))
	} else {
		q, _ = siftEquel(q, "src", spec.Src, false)
	}
	q, _ = spec.TextSearchSpec.Sift(q)

	return q
}
func (spec *ArticleSpec) CanSort(k string) bool {
	switch k {
	case "author", "news_publish":
		return true
	default:
		return spec.ModelSpec.CanSort(k)
	}
}

type AttachmentSpec struct {
	comm.PageSpec
	ModelSpec

	// 文章编号
	ArticleID string `extensions:"x-order=A" form:"articleID" json:"articleID"`
	// 名称
	Name string `extensions:"x-order=B" form:"name" json:"name"`
	// 类型
	Mime string `extensions:"x-order=C" form:"mime" json:"mime"`
	Path string `extensions:"x-order=D" form:"path" json:"path"`
}

func (spec *AttachmentSpec) Sift(q *ormQuery) *ormQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = siftOID(q, "article_id", spec.ArticleID, false)
	q, _ = siftMatch(q, "name", spec.Name, false)
	q, _ = siftICE(q, "mime", spec.Mime, false)
	q, _ = siftMatch(q, "path", spec.Path, false)

	return q
}

type contentStore struct {
	w *Wrap
}

func (s *contentStore) ListClause(ctx context.Context, spec *ClauseSpec) (data cms1.Clauses, total int, err error) {
	total, err = queryPager(ctx, spec, s.w.db.NewSelect().Model(&data).Apply(spec.Sift))
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
	spec.SetTsConfig(s.w.db.GetTsCfg())
	spec.SetTsFallback("title", "content")
	total, err = queryPager(ctx, spec, s.w.db.NewSelect().Model(&data).Apply(spec.Sift))
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
		obj.SetTsColumns("title", "content")
	}
	err = s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
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
	s.w.opModelMeta(ctx, exist)
	if tscfg, ok := s.w.db.GetTsCfg(); ok {
		exist.TsCfgName = tscfg
		exist.SetTsColumns("title", "content")
	} else {
		exist.TsCfgName = ""
	}
	exist.SetChange("ts_cfg")
	return s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
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
	return s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		err = dbDeleteT(ctx, tx, s.w.db.Schema(), s.w.db.SchemaCrap(), "cms_article", obj.ID)
		if err != nil {
			return
		}
		return dbAfterDeleteArticle(ctx, tx, obj)
	})
}

func (s *contentStore) ListAttachment(ctx context.Context, spec *AttachmentSpec) (data cms1.Attachments, total int, err error) {
	total, err = queryPager(ctx, spec, s.w.db.NewSelect().Model(&data).Apply(spec.Sift))
	return
}
func (s *contentStore) GetAttachment(ctx context.Context, id string) (obj *cms1.Attachment, err error) {
	obj = new(cms1.Attachment)
	err = getModelWithPKID(ctx, s.w.db, obj, id)
	return
}
