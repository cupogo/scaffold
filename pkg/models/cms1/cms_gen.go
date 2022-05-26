// This file is generated - Do Not Edit.

package cms1

import (
	comm "hyyl.xyz/cupola/aurora/pkg/models/comm"
	oid "hyyl.xyz/cupola/aurora/pkg/models/oid"
)

// Article 文章
type Article struct {
	tableName struct{} `pg:"cms_article,alias:a"`

	comm.DefaultModel

	ArticleBasic
}

type ArticleBasic struct {
	Author  string `json:"author" pg:",notnull"`  // 作者
	Title   string `json:"title" pg:",notnull"`   // 标题
	Contant string `json:"content" pg:",notnull"` // 内容
}

type Articles []Article

// Creating function call to it's inner fields defined hooks
func (z *Article) Creating() error {
	if z.ID.IsZero() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}

// Saving function call to it's inner fields defined hooks
func (z *Article) Saving() error {
	return z.DefaultModel.Creating()
}

type ArticleSet struct {
	Author  *string `json:"author"`  // 作者
	Title   *string `json:"title"`   // 标题
	Contant *string `json:"content"` // 内容
}

func (z *Article) SetWith(o *ArticleSet) (cs []string) {
	if o.Author != nil {
		z.Author = *o.Author
		cs = append(cs, "author")
	}
	if o.Title != nil {
		z.Title = *o.Title
		cs = append(cs, "title")
	}
	if o.Contant != nil {
		z.Contant = *o.Contant
		cs = append(cs, "contant")
	}
	return
}

// Clause 条款
type Clause struct {
	tableName struct{} `pg:"cms_clause,alias:c"`

	comm.DefaultModel

	ClauseBasic
}

type ClauseBasic struct {
	Text string `json:"text" pg:"text,notnull"`
}

type Clauses []Clause

// Creating function call to it's inner fields defined hooks
func (z *Clause) Creating() error {
	if z.ID.IsZero() {
		z.SetID(oid.NewID(oid.OtArticle))
	}

	return z.DefaultModel.Creating()
}

// Saving function call to it's inner fields defined hooks
func (z *Clause) Saving() error {
	return z.DefaultModel.Creating()
}

type ClauseSet struct {
	Text *string `json:"text"`
}

func (z *Clause) SetWith(o *ClauseSet) (cs []string) {
	if o.Text != nil {
		z.Text = *o.Text
		cs = append(cs, "text")
	}
	return
}
