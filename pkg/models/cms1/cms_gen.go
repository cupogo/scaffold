// This file is generated - Do Not Edit.

package cms1

import (
	comm "github.com/cupogo/andvari/models/comm"
	oid "github.com/cupogo/andvari/models/oid"
)

// consts of Channel 频道
const (
	ChannelTable = "cms_channel"
	ChannelAlias = "c"
	ChannelLabel = "channel"
	ChannelModel = "cms1Channel"
)

// Channel 频道
type Channel struct {
	comm.BaseModel `bun:"table:cms_channel,alias:c" json:"-"`

	comm.DefaultModel

	ChannelBasic

	comm.MetaField
} // @name cms1Channel

type ChannelBasic struct {
	// 自定义短ID
	Slug string `bun:"slug,notnull,type:name,unique" extensions:"x-order=A" form:"slug" json:"key" pg:"slug,notnull,type:name,unique"`
	// 父级ID
	ParentID oid.OID `bun:",notnull" extensions:"x-order=B" json:"parentID" pg:",notnull,use_zero" swaggertype:"string"`
	// 名称
	Name string `bun:",notnull" extensions:"x-order=C" form:"name" json:"name" pg:",notnull"`
	// 描述
	Description string `bun:",notnull" extensions:"x-order=D" form:"description" json:"description,omitempty" pg:",notnull,use_zero"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" bun:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
} // @name cms1ChannelBasic

type Channels []Channel

// Creating function call to it's inner fields defined hooks
func (z *Channel) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}
func NewChannelWithBasic(in ChannelBasic) *Channel {
	obj := &Channel{
		ChannelBasic: in,
	}
	_ = obj.MetaUp(in.MetaDiff)
	return obj
}
func NewChannelWithID(id any) *Channel {
	obj := new(Channel)
	_ = obj.SetID(id)
	return obj
}
func (_ *Channel) IdentityLabel() string { return ChannelLabel }
func (_ *Channel) IdentityModel() string { return ChannelModel }
func (_ *Channel) IdentityTable() string { return ChannelTable }
func (_ *Channel) IdentityAlias() string { return ChannelAlias }

type ChannelSet struct {
	// 自定义短ID
	Slug *string `extensions:"x-order=A" form:"slug" json:"key"`
	// 父级ID
	ParentID *string `extensions:"x-order=B" json:"parentID"`
	// 名称
	Name *string `extensions:"x-order=C" json:"name"`
	// 描述
	Description *string `extensions:"x-order=D" form:"description" json:"description,omitempty"`
	// for meta update
	MetaDiff *comm.MetaDiff `json:"metaUp,omitempty" swaggerignore:"true"`
} // @name cms1ChannelSet

func (z *Channel) SetWith(o ChannelSet) {
	if o.Slug != nil && z.Slug != *o.Slug {
		z.LogChangeValue("slug", z.Slug, o.Slug)
		z.Slug = *o.Slug
	}
	if o.ParentID != nil {
		if id := oid.Cast(*o.ParentID); z.ParentID != id {
			z.LogChangeValue("parent_id", z.ParentID, id)
			z.ParentID = id
		}
	}
	if o.Name != nil && z.Name != *o.Name {
		z.LogChangeValue("name", z.Name, o.Name)
		z.Name = *o.Name
	}
	if o.Description != nil && z.Description != *o.Description {
		z.LogChangeValue("description", z.Description, o.Description)
		z.Description = *o.Description
	}
	if o.MetaDiff != nil && z.MetaUp(o.MetaDiff) {
		z.SetChange("meta")
	}
}
func (in *ChannelBasic) MetaAddKVs(args ...any) *ChannelBasic {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}
func (in *ChannelSet) MetaAddKVs(args ...any) *ChannelSet {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}

// consts of Article 文章
const (
	ArticleTable = "cms_article"
	ArticleAlias = "a"
	ArticleLabel = "article"
	ArticleModel = "cms1Article"
)

// Article 文章
// @Description 文章示例
// @Description 有关说明
type Article struct {
	comm.BaseModel `bun:"table:cms_article,alias:a" json:"-"`

	comm.DefaultModel

	ArticleBasic

	comm.MetaField

	comm.TextSearchField
} // @name cms1Article

type ArticleBasic struct {
	// 作者
	Author string `bun:",notnull" extensions:"x-order=A" form:"author" json:"author" pg:",notnull,use_zero"`
	// 标题
	Title string `bun:",notnull" extensions:"x-order=B" form:"title" json:"title" pg:",notnull"`
	// 内容
	Content string `bun:",notnull" extensions:"x-order=C" form:"content" json:"content" pg:",notnull"`
	// 新闻时间
	NewsPublish comm.DateTime `bun:"news_publish,type:date" extensions:"x-order=D" form:"newsPublish" json:"newsPublish,omitempty" pg:"news_publish,type:date"`
	// 状态
	Status int16 `bun:",notnull" extensions:"x-order=E" form:"status" json:"status" pg:",notnull,use_zero"`
	// 作者编号
	AuthorID oid.OID `bun:",notnull" extensions:"x-order=F" json:"authorID" pg:",notnull,use_zero" swaggertype:"string"`
	// 来源
	Src string `bun:",notnull" extensions:"x-order=G" form:"src" json:"src" pg:",notnull,use_zero"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" bun:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
} // @name cms1ArticleBasic

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
func NewArticleWithID(id any) *Article {
	obj := new(Article)
	_ = obj.SetID(id)
	return obj
}
func (_ *Article) IdentityLabel() string { return ArticleLabel }
func (_ *Article) IdentityModel() string { return ArticleModel }
func (_ *Article) IdentityTable() string { return ArticleTable }
func (_ *Article) IdentityAlias() string { return ArticleAlias }
func (_ *Article) WithFK() bool {
	return true
}

type ArticleSet struct {
	// 作者
	Author *string `extensions:"x-order=A" json:"author"`
	// 标题
	Title *string `extensions:"x-order=B" json:"title"`
	// 内容
	Content *string `extensions:"x-order=C" json:"content"`
	// 新闻时间
	NewsPublish *comm.DateTime `extensions:"x-order=D" form:"newsPublish" json:"newsPublish,omitempty"`
	// 状态
	Status *int16 `extensions:"x-order=E" json:"status"`
	// 作者编号
	AuthorID *string `extensions:"x-order=F" json:"authorID"`
	// 来源
	Src *string `extensions:"x-order=G" json:"src"`
	// for meta update
	MetaDiff *comm.MetaDiff `json:"metaUp,omitempty" swaggerignore:"true"`
} // @name cms1ArticleSet

func (z *Article) SetWith(o ArticleSet) {
	if o.Author != nil && z.Author != *o.Author {
		z.LogChangeValue("author", z.Author, o.Author)
		z.Author = *o.Author
	}
	if o.Title != nil && z.Title != *o.Title {
		z.LogChangeValue("title", z.Title, o.Title)
		z.Title = *o.Title
	}
	if o.Content != nil && z.Content != *o.Content {
		z.LogChangeValue("content", z.Content, o.Content)
		z.Content = *o.Content
	}
	if o.NewsPublish != nil && z.NewsPublish != *o.NewsPublish {
		z.LogChangeValue("news_publish", z.NewsPublish, o.NewsPublish)
		z.NewsPublish = *o.NewsPublish
	}
	if o.Status != nil && z.Status != *o.Status {
		z.LogChangeValue("status", z.Status, o.Status)
		z.Status = *o.Status
	}
	if o.AuthorID != nil {
		if id := oid.Cast(*o.AuthorID); z.AuthorID != id {
			z.LogChangeValue("author_id", z.AuthorID, id)
			z.AuthorID = id
		}
	}
	if o.Src != nil && z.Src != *o.Src {
		z.LogChangeValue("src", z.Src, o.Src)
		z.Src = *o.Src
	}
	if o.MetaDiff != nil && z.MetaUp(o.MetaDiff) {
		z.SetChange("meta")
	}
}
func (in *ArticleBasic) MetaAddKVs(args ...any) *ArticleBasic {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}
func (in *ArticleSet) MetaAddKVs(args ...any) *ArticleSet {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}

// consts of Attachment 附件
const (
	AttachmentTable = "cms_attachment"
	AttachmentAlias = "att"
	AttachmentLabel = "attachment"
	AttachmentModel = "cms1Attachment"
)

// Attachment 附件
type Attachment struct {
	comm.BaseModel `bun:"table:cms_attachment,alias:att" json:"-"`

	comm.DefaultModel

	AttachmentBasic

	comm.MetaField
} // @name cms1Attachment

type AttachmentBasic struct {
	// 文章编号
	ArticleID oid.OID `bun:",notnull" extensions:"x-order=A" json:"articleID" pg:",notnull" swaggertype:"string"`
	// 名称
	Name string `bun:",notnull" extensions:"x-order=B" form:"name" json:"name" pg:",notnull"`
	// 类型
	Mime string `bun:",notnull" extensions:"x-order=C" form:"mime" json:"mime" pg:",notnull"`
	Path string `bun:"path,notnull" extensions:"x-order=D" form:"path" json:"path" pg:"path,notnull"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" bun:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
} // @name cms1AttachmentBasic

type Attachments []Attachment

// Creating function call to it's inner fields defined hooks
func (z *Attachment) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtFile))
	}

	return z.DefaultModel.Creating()
}
func NewAttachmentWithBasic(in AttachmentBasic) *Attachment {
	obj := &Attachment{
		AttachmentBasic: in,
	}
	_ = obj.MetaUp(in.MetaDiff)
	return obj
}
func NewAttachmentWithID(id any) *Attachment {
	obj := new(Attachment)
	_ = obj.SetID(id)
	return obj
}
func (_ *Attachment) IdentityLabel() string { return AttachmentLabel }
func (_ *Attachment) IdentityModel() string { return AttachmentModel }
func (_ *Attachment) IdentityTable() string { return AttachmentTable }
func (_ *Attachment) IdentityAlias() string { return AttachmentAlias }

type AttachmentSet struct {
	// 文章编号
	ArticleID *string `extensions:"x-order=A" json:"articleID"`
	// 名称
	Name *string `extensions:"x-order=B" json:"name"`
	// 类型
	Mime *string `extensions:"x-order=C" json:"mime"`
	Path *string `extensions:"x-order=D" json:"path"`
	// for meta update
	MetaDiff *comm.MetaDiff `json:"metaUp,omitempty" swaggerignore:"true"`
} // @name cms1AttachmentSet

func (z *Attachment) SetWith(o AttachmentSet) {
	if o.ArticleID != nil {
		if id := oid.Cast(*o.ArticleID); z.ArticleID != id {
			z.LogChangeValue("article_id", z.ArticleID, id)
			z.ArticleID = id
		}
	}
	if o.Name != nil && z.Name != *o.Name {
		z.LogChangeValue("name", z.Name, o.Name)
		z.Name = *o.Name
	}
	if o.Mime != nil && z.Mime != *o.Mime {
		z.LogChangeValue("mime", z.Mime, o.Mime)
		z.Mime = *o.Mime
	}
	if o.Path != nil && z.Path != *o.Path {
		z.LogChangeValue("path", z.Path, o.Path)
		z.Path = *o.Path
	}
	if o.MetaDiff != nil && z.MetaUp(o.MetaDiff) {
		z.SetChange("meta")
	}
}
func (in *AttachmentBasic) MetaAddKVs(args ...any) *AttachmentBasic {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}
func (in *AttachmentSet) MetaAddKVs(args ...any) *AttachmentSet {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}

// consts of Clause 条款
const (
	ClauseTable = "cms_clause"
	ClauseAlias = "c"
	ClauseLabel = "clause"
	ClauseModel = "cms1Clause"
)

// Clause 条款
type Clause struct {
	comm.BaseModel `bun:"table:cms_clause,alias:c" json:"-"`

	comm.DefaultModel

	ClauseBasic
} // @name cms1Clause

type ClauseBasic struct {
	Text string `bun:"text,notnull" extensions:"x-order=A" form:"text" json:"text" pg:"text,notnull"`
} // @name cms1ClauseBasic

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
func NewClauseWithID(id any) *Clause {
	obj := new(Clause)
	_ = obj.SetID(id)
	return obj
}
func (_ *Clause) IdentityLabel() string { return ClauseLabel }
func (_ *Clause) IdentityModel() string { return ClauseModel }
func (_ *Clause) IdentityTable() string { return ClauseTable }
func (_ *Clause) IdentityAlias() string { return ClauseAlias }

type ClauseSet struct {
	Text *string `extensions:"x-order=A" json:"text"`
} // @name cms1ClauseSet

func (z *Clause) SetWith(o ClauseSet) {
	if o.Text != nil && z.Text != *o.Text {
		z.LogChangeValue("text", z.Text, o.Text)
		z.Text = *o.Text
	}
}

// consts of File a
const (
	FileLabel = "file"
	FileModel = "cms1File"
)

// File a file instance
type File struct {
	Name string `extensions:"x-order=A" form:"name" json:"name"`
	Path string `extensions:"x-order=B" form:"path" json:"path"`
} // @name cms1File
