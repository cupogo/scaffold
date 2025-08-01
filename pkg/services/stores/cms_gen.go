// This file is generated - Do Not Edit.

package stores

import (
	"context"
	"fmt"

	pgx "github.com/cupogo/andvari/stores/pgx"
	utils "github.com/cupogo/andvari/utils"
	"github.com/cupogo/scaffold/pkg/models/cms1"
)

// type Article = cms1.Article
// type Attachment = cms1.Attachment
// type Channel = cms1.Channel
// type Clause = cms1.Clause
// type File = cms1.File

func init() {
	RegisterModel((*cms1.Channel)(nil), (*cms1.Article)(nil), (*cms1.Attachment)(nil), (*cms1.Clause)(nil))
}

type ContentStore interface {
	ListClause(ctx context.Context, spec *ClauseSpec) (data cms1.Clauses, total int, err error)
	GetClause(ctx context.Context, id string) (obj *cms1.Clause, err error)
	PutClause(ctx context.Context, id string, in cms1.ClauseSet) (obj *cms1.Clause, err error)
	DeleteClause(ctx context.Context, id string) error

	ListChannel(ctx context.Context, spec *ChannelSpec) (data cms1.Channels, total int, err error)
	GetChannel(ctx context.Context, id string) (obj *cms1.Channel, err error)
	PutChannel(ctx context.Context, id string, in cms1.ChannelSet) (obj *cms1.Channel, err error)
	DeleteChannel(ctx context.Context, id string) error

	ListArticle(ctx context.Context, spec *ArticleSpec) (data cms1.Articles, total int, err error)
	GetArticle(ctx context.Context, id string) (obj *cms1.Article, err error)
	CreateArticle(ctx context.Context, in cms1.ArticleBasic) (obj *cms1.Article, err error)
	UpdateArticle(ctx context.Context, id string, in cms1.ArticleSet) error
	DeleteArticle(ctx context.Context, id string) error

	ListAttachment(ctx context.Context, spec *AttachmentSpec) (data cms1.Attachments, total int, err error)
	GetAttachment(ctx context.Context, id string) (obj *cms1.Attachment, err error)
	CreateAttachment(ctx context.Context, in cms1.AttachmentBasic) (obj *cms1.Attachment, err error)
	DeleteAttachment(ctx context.Context, id string) error
}

type ClauseSpec struct {
	PageSpec
	ModelSpec

	Text string `extensions:"x-order=A" form:"text" json:"text"`
}

func (spec *ClauseSpec) Sift(q *ormQuery) *ormQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = siftMatch(q, "text", spec.Text, false)

	return q
}

type ChannelSpec struct {
	PageSpec
	ModelSpec

	// 自定义短ID
	Slug string `extensions:"x-order=A" form:"slug" json:"key"`
	// 父级ID
	ParentID string `extensions:"x-order=B" form:"parentID" json:"parentID"`
	// 名称
	Name string `extensions:"x-order=C" form:"name" json:"name"`
}

func (spec *ChannelSpec) Sift(q *ormQuery) *ormQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = siftEqual(q, "slug", spec.Slug, false)
	q, _ = siftOID(q, "parent_id", spec.ParentID, false)
	q, _ = siftMatch(q, "name", spec.Name, false, true)

	return q
}

type ArticleSpec struct {
	PageSpec
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
	// 作者编号
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
		q, _ = sift(q, "status", "IN", vals, false)
	} else {
		q, _ = siftEqual(q, "status", spec.Status, false)
	}
	q, _ = siftOIDs(q, "author_id", spec.AuthorID, false)
	if vals, ok := utils.ParseStrs(spec.Srcs); ok {
		q, _ = sift(q, "src", "IN", vals, false)
	} else {
		q, _ = siftEqual(q, "src", spec.Src, false)
	}
	q = spec.TextSearchSpec.SiftTS(q, !spec.HasColumn())

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
	PageSpec
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

func newContentStore(w *Wrap) *contentStore {
	s := &contentStore{w: w}
	RegisterESMigrate((*cms1.Article)(nil), s.MigrateESArticle)
	return s
}

type contentStore struct {
	w *Wrap
}

func (s *contentStore) ListClause(ctx context.Context, spec *ClauseSpec) (data cms1.Clauses, total int, err error) {
	total, err = s.w.db.ListModel(ctx, spec, &data)
	return
}
func (s *contentStore) GetClause(ctx context.Context, id string) (obj *cms1.Clause, err error) {
	obj = new(cms1.Clause)
	err = dbGetWithPKID(ctx, s.w.db, obj, id)

	return
}
func (s *contentStore) PutClause(ctx context.Context, id string, in cms1.ClauseSet) (obj *cms1.Clause, err error) {
	obj, err = pgx.StoreWithSet[*cms1.Clause](ctx, s.w.db, in, id)
	return
}
func (s *contentStore) DeleteClause(ctx context.Context, id string) error {
	obj := new(cms1.Clause)
	return s.w.db.DeleteModel(ctx, obj, id)
}

func (s *contentStore) ListChannel(ctx context.Context, spec *ChannelSpec) (data cms1.Channels, total int, err error) {
	total, err = s.w.db.ListModel(ctx, spec, &data)
	return
}
func (s *contentStore) GetChannel(ctx context.Context, id string) (obj *cms1.Channel, err error) {
	obj = new(cms1.Channel)
	if err = dbGetWith(ctx, s.w.db, obj, "slug", "=", id); err != nil {
		err = dbGetWithPKID(ctx, s.w.db, obj, id)
	}

	return
}
func (s *contentStore) PutChannel(ctx context.Context, id string, in cms1.ChannelSet) (obj *cms1.Channel, err error) {
	if in.Slug == nil || *in.Slug == "" {
		err = fmt.Errorf("need slug")
		return
	}
	if len(id) > 0 {
		obj, err = pgx.StoreWithSet[*cms1.Channel](ctx, s.w.db, in, id)
	} else {
		obj, err = pgx.StoreWithSet[*cms1.Channel](ctx, s.w.db, in, *in.Slug, "slug")
	}
	return
}
func (s *contentStore) DeleteChannel(ctx context.Context, id string) error {
	obj := new(cms1.Channel)
	return s.w.db.DeleteModel(ctx, obj, id)
}

func (s *contentStore) ListArticle(ctx context.Context, spec *ArticleSpec) (data cms1.Articles, total int, err error) {
	spec.SetTsConfig(s.w.db.GetTsCfg())
	spec.SetTsFallback("title", "content")
	q := s.w.db.NewSelect().Model(&data)
	if err = s.beforeListArticle(ctx, spec, q); err != nil {
		return
	}
	total, err = queryPager(ctx, spec, q)
	if err == nil {
		err = s.afterListArticle(ctx, spec, data)
	}
	return
}
func (s *contentStore) GetArticle(ctx context.Context, id string) (obj *cms1.Article, err error) {
	obj = new(cms1.Article)
	err = dbGetWithPKID(ctx, s.w.db, obj, id)
	if err == nil {
		err = s.afterLoadArticle(ctx, obj)
	}
	return
}
func (s *contentStore) CreateArticle(ctx context.Context, in cms1.ArticleBasic) (obj *cms1.Article, err error) {
	err = s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		obj, err = CreateArticle(ctx, tx, in)
		return err
	})
	if err == nil {
		err = s.upsertESArticle(ctx, obj)
	}
	return
}
func (s *contentStore) UpdateArticle(ctx context.Context, id string, in cms1.ArticleSet) error {
	var exist *cms1.Article
	if err := s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		exist, err = UpdateArticle(ctx, tx, id, in)
		return err
	}); err != nil {
		return err
	}
	return s.upsertESArticle(ctx, exist)
}
func (s *contentStore) DeleteArticle(ctx context.Context, id string) error {
	obj := new(cms1.Article)
	if err := dbGetWithPKID(ctx, s.w.db, obj, id); err != nil {
		return err
	}
	if err := s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		err = dbDeleteM(ctx, tx, s.w.db.Schema(), s.w.db.SchemaCrap(), obj)
		if err != nil {
			return
		}
		return dbAfterDeleteArticle(ctx, tx, obj)
	}); err != nil {
		return err
	}
	return s.deleteESArticle(ctx, obj)
}

func (s *contentStore) ListAttachment(ctx context.Context, spec *AttachmentSpec) (data cms1.Attachments, total int, err error) {
	total, err = s.w.db.ListModel(ctx, spec, &data)
	return
}
func (s *contentStore) GetAttachment(ctx context.Context, id string) (obj *cms1.Attachment, err error) {
	obj = new(cms1.Attachment)
	err = dbGetWithPKID(ctx, s.w.db, obj, id)

	return
}
func (s *contentStore) CreateAttachment(ctx context.Context, in cms1.AttachmentBasic) (obj *cms1.Attachment, err error) {
	obj, err = CreateAttachment(ctx, s.w.db, in)
	return
}
func (s *contentStore) DeleteAttachment(ctx context.Context, id string) error {
	obj := new(cms1.Attachment)
	return s.w.db.DeleteModel(ctx, obj, id)
}

func CreateArticle(ctx context.Context, db ormDB, in cms1.ArticleBasic) (obj *cms1.Article, err error) {
	obj = cms1.NewArticleWithBasic(in)
	if tscfg, ok := DbTsCheck(); ok {
		obj.TsCfgName = tscfg
		obj.SetTsColumns("title", "content")
	}
	if err = dbBeforeSaveArticle(ctx, db, obj); err != nil {
		return
	}
	dbMetaUp(ctx, db, obj)
	err = dbInsert(ctx, db, obj)
	if err == nil {
		err = dbAfterCreateArticle(ctx, db, obj)
	}
	return
}
func UpdateArticle(ctx context.Context, db ormDB, id string, in cms1.ArticleSet) (exist *cms1.Article, err error) {
	exist = new(cms1.Article)
	if err = dbGetWithPKID(ctx, db, exist, id); err != nil {
		return
	}
	exist.SetIsUpdate(true)
	exist.SetWith(in)
	if tscfg, ok := DbTsCheck(); ok {
		exist.TsCfgName = tscfg
		exist.SetTsColumns("title", "content")
		exist.SetChange("ts_cfg")
	}
	if err = dbBeforeSaveArticle(ctx, db, exist); err != nil {
		return
	}
	dbMetaUp(ctx, db, exist)
	if err = dbUpdate(ctx, db, exist); err != nil {
		return
	}
	err = dbAfterUpdateArticle(ctx, db, exist)
	return
}
func CreateAttachment(ctx context.Context, db ormDB, in cms1.AttachmentBasic) (obj *cms1.Attachment, err error) {
	obj = cms1.NewAttachmentWithBasic(in)
	dbMetaUp(ctx, db, obj)
	err = dbInsert(ctx, db, obj)
	return
}
