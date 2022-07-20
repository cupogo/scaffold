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
} // @name Article

type ArticleBasic struct {
	// 作者
	Author string `json:"author" pg:",notnull"`
	// 标题
	Title string `json:"title" pg:",notnull"`
	// 内容
	Content string `json:"content" pg:",notnull"`
} // @name ArticleBasic

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
	return z.DefaultModel.Saving()
}

type ArticleSet struct {
	Author  *string `json:"author"`  // 作者
	Title   *string `json:"title"`   // 标题
	Content *string `json:"content"` // 内容
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
	return
}

// Clause 条款
type Clause struct {
	tableName struct{} `pg:"cms_clause,alias:c"`

	comm.DefaultModel

	ClauseBasic
} // @name Clause

type ClauseBasic struct {
	Text string `json:"text" pg:"text,notnull"`
} // @name ClauseBasic

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
	return z.DefaultModel.Saving()
}

type ClauseSet struct {
	Text *string `json:"text"`
} // @name ClauseSet

func (z *Clause) SetWith(o ClauseSet) (cs []string) {
	if o.Text != nil {
		z.Text = *o.Text
		cs = append(cs, "text")
	}
	return
}
