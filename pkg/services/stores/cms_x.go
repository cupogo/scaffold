package stores

import (
	"context"

	"github.com/cupogo/scaffold/pkg/models/cms1"
)

func dbBeforeSaveArticle(ctx context.Context, db ormDB, obj *cms1.Article) error {
	// TODO:
	return nil
}
func dbAfterDeleteArticle(ctx context.Context, db ormDB, obj *cms1.Article) error {
	// TODO:
	return nil
}
func (s *contentStore) beforeListArticle(ctx context.Context, spec *ArticleSpec, q *ormQuery) error {
	// TODO:
	return nil
}
func (s *contentStore) afterLoadArticle(ctx context.Context, obj *cms1.Article) error {
	// TODO:
	return nil
}
func (s *contentStore) afterListArticle(ctx context.Context, spec *ArticleSpec, data cms1.Articles) error {
	// TODO:
	return nil
}
