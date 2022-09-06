// This file is generated - Do Not Edit.

package cms1

import (
	comm "hyyl.xyz/cupola/andvari/models/comm"
	oid "hyyl.xyz/cupola/andvari/models/oid"
)

// Article 文章
type Article struct {
	tableName struct{} `pg:"cms_article,alias:a"`

	comm.DefaultModel

	ArticleBasic

	comm.MetaField

	comm.TextSearchField
} // @name Article

type ArticleBasic struct {
	// 作者
	Author string `extensions:"x-order=A" form:"author" json:"author" pg:",notnull,use_zero"`
	// 标题
	Title string `extensions:"x-order=B" form:"title" json:"title" pg:",notnull"`
	// 内容
	Content string `extensions:"x-order=C" form:"content" json:"content" pg:",notnull"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
} // @name ArticleBasic

type Articles []Article

// Creating function call to it's inner fields defined hooks
func (z *Article) Creating() error {
	if z.ID.IsZero() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}

type ArticleSet struct {
	// 作者
	Author *string `extensions:"x-order=A" json:"author"`
	// 标题
	Title *string `extensions:"x-order=B" json:"title"`
	// 内容
	Content *string `extensions:"x-order=C" json:"content"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
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
	if o.MetaDiff != nil && z.MetaUp(o.MetaDiff) {
		cs = append(cs, "meta")
	}
	if len(cs) > 0 {
		z.SetChange(cs...)
	}
	return
}

// Clause 条款
type Clause struct {
	tableName struct{} `pg:"cms_clause,alias:c"`

	comm.DefaultModel

	ClauseBasic
} // @name Clause

type ClauseBasic struct {
	Text string `extensions:"x-order=A" form:"text" json:"text" pg:"text,notnull"`
} // @name ClauseBasic

type Clauses []Clause

// Creating function call to it's inner fields defined hooks
func (z *Clause) Creating() error {
	if z.ID.IsZero() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
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
