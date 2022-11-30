// This file is generated - Do Not Edit.

package cms1

import (
	comm "github.com/cupogo/andvari/models/comm"
	oid "github.com/cupogo/andvari/models/oid"
)

// table name of Article 文章
const TableArticle = "cms_article"
const TableArticleAlias = "a"

// Article 文章
type Article struct {
	comm.BaseModel `bun:"table:cms_article,alias:a" json:"-"`

	comm.DefaultModel

	ArticleBasic

	comm.MetaField

	comm.TextSearchField
} // @name Article

type ArticleBasic struct {
	// 作者
	Author string `bun:",notnull" extensions:"x-order=A" form:"author" json:"author" pg:",notnull,use_zero"`
	// 标题
	Title string `bun:",notnull" extensions:"x-order=B" form:"title" json:"title" pg:",notnull"`
	// 内容
	Content string `bun:",notnull" extensions:"x-order=C" form:"content" json:"content" pg:",notnull"`
	// 新闻时间
	NewsPublish comm.DateTime `bun:"news_publish,type:date" extensions:"x-order=D" json:"newsPublish,omitempty" pg:"news_publish,type:date"`
	// 状态
	Status int16 `bun:",notnull" extensions:"x-order=E" form:"status" json:"status" pg:",notnull,use_zero"`
	// 作者
	AuthorID oid.OID `bun:",notnull" extensions:"x-order=F" json:"authorID" pg:",notnull,use_zero"`
	// 来源
	Src string `bun:",notnull" extensions:"x-order=G" form:"src" json:"src" pg:",notnull,use_zero"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" bun:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
} // @name ArticleBasic

type Articles []Article

// Creating function call to it's inner fields defined hooks
func (z *Article) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}
func NewArticleWithBasic(in ArticleBasic) *Article {
	obj := &Article{
		ArticleBasic: in,
	}
	_ = obj.MetaUp(in.MetaDiff)
	return obj
}

type ArticleSet struct {
	// 作者
	Author *string `extensions:"x-order=A" json:"author"`
	// 标题
	Title *string `extensions:"x-order=B" json:"title"`
	// 内容
	Content *string `extensions:"x-order=C" json:"content"`
	// 新闻时间
	NewsPublish *comm.DateTime `extensions:"x-order=D" json:"newsPublish,omitempty"`
	// 状态
	Status *int16 `extensions:"x-order=E" json:"status"`
	// 作者
	AuthorID *string `extensions:"x-order=F" json:"authorID"`
	// 来源
	Src *string `extensions:"x-order=G" json:"src"`
	// for meta update
	MetaDiff *comm.MetaDiff `json:"metaUp,omitempty" swaggerignore:"true"`
} // @name ArticleSet

func (z *Article) SetWith(o ArticleSet) (cs []string) {
	if o.Author != nil {
		z.Author = *o.Author
		cs = append(cs, "author")
	}
	if o.Title != nil {
		z.Title = *o.Title
		cs = append(cs, "title")
	}
	if o.Content != nil {
		z.Content = *o.Content
		cs = append(cs, "content")
	}
	if o.NewsPublish != nil {
		z.NewsPublish = *o.NewsPublish
		cs = append(cs, "news_publish")
	}
	if o.Status != nil {
		z.Status = *o.Status
		cs = append(cs, "status")
	}
	if o.AuthorID != nil {
		z.AuthorID = oid.Cast(*o.AuthorID)
		cs = append(cs, "author_id")
	}
	if o.Src != nil {
		z.Src = *o.Src
		cs = append(cs, "src")
	}
	if o.MetaDiff != nil && z.MetaUp(o.MetaDiff) {
		cs = append(cs, "meta")
	}
	if len(cs) > 0 {
		z.SetChange(cs...)
	}
	return
}

// table name of Attachment 附件
const TableAttachment = "cms_attachment"
const TableAttachmentAlias = "att"

// Attachment 附件
type Attachment struct {
	comm.BaseModel `bun:"table:cms_attachment,alias:att" json:"-"`

	comm.DefaultModel

	AttachmentBasic
} // @name Attachment

type AttachmentBasic struct {
	// 文章编号
	ArticleID oid.OID `bun:",notnull" extensions:"x-order=A" json:"articleID" pg:",notnull"`
	// 名称
	Name string `bun:",notnull" extensions:"x-order=B" form:"name" json:"name" pg:",notnull"`
	// 类型
	Mime string `bun:",notnull" extensions:"x-order=C" form:"mime" json:"mime" pg:",notnull"`
	Path string `bun:"path,notnull" extensions:"x-order=D" form:"path" json:"path" pg:"path,notnull"`
} // @name AttachmentBasic

type Attachments []Attachment

// Creating function call to it's inner fields defined hooks
func (z *Attachment) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}
func NewAttachmentWithBasic(in AttachmentBasic) *Attachment {
	obj := &Attachment{
		AttachmentBasic: in,
	}
	return obj
}

type AttachmentSet struct {
	// 文章编号
	ArticleID *string `extensions:"x-order=A" json:"articleID"`
	// 名称
	Name *string `extensions:"x-order=B" json:"name"`
	// 类型
	Mime *string `extensions:"x-order=C" json:"mime"`
	Path *string `extensions:"x-order=D" json:"path"`
} // @name AttachmentSet

func (z *Attachment) SetWith(o AttachmentSet) (cs []string) {
	if o.ArticleID != nil {
		z.ArticleID = oid.Cast(*o.ArticleID)
		cs = append(cs, "article_id")
	}
	if o.Name != nil {
		z.Name = *o.Name
		cs = append(cs, "name")
	}
	if o.Mime != nil {
		z.Mime = *o.Mime
		cs = append(cs, "mime")
	}
	if o.Path != nil {
		z.Path = *o.Path
		cs = append(cs, "path")
	}
	if len(cs) > 0 {
		z.SetChange(cs...)
	}
	return
}

// table name of Clause 条款
const TableClause = "cms_clause"
const TableClauseAlias = "c"

// Clause 条款
type Clause struct {
	comm.BaseModel `bun:"table:cms_clause,alias:c" json:"-"`

	comm.DefaultModel

	ClauseBasic
} // @name Clause

type ClauseBasic struct {
	Text string `bun:"text,notnull" extensions:"x-order=A" form:"text" json:"text" pg:"text,notnull"`
} // @name ClauseBasic

type Clauses []Clause

// Creating function call to it's inner fields defined hooks
func (z *Clause) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}
func NewClauseWithBasic(in ClauseBasic) *Clause {
	obj := &Clause{
		ClauseBasic: in,
	}
	return obj
}

type ClauseSet struct {
	Text *string `extensions:"x-order=A" json:"text"`
} // @name ClauseSet

func (z *Clause) SetWith(o ClauseSet) (cs []string) {
	if o.Text != nil {
		z.Text = *o.Text
		cs = append(cs, "text")
	}
	if len(cs) > 0 {
		z.SetChange(cs...)
	}
	return
}

// File a file instance
type File struct {
	Name string `extensions:"x-order=A" form:"name" json:"name"`
	Path string `extensions:"x-order=B" form:"path" json:"path"`
} // @name File

type Files []File
