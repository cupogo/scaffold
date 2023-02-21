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
func (s *contentStore) upsertESArticle(ctx context.Context, obj *cms1.Article) error {
	if obj == nil {
		return nil
	}
	// TODO:
	err := UpsertESDoc(ctx, obj.IdentityTable(), obj)
	if err != nil {
		logger().Infow("UpsertESDoc", "index", obj.IdentityTable(), "error", err)
		return nil
	}
	return nil
}
func (s *contentStore) MigrateESArticle(ctx context.Context) (err error) {
	var (
		ms          cms1.Articles
		limit, page = 1000, 1
	)
	for {
		var spec ArticleSpec
		spec.Limit = limit
		spec.Page = page
		spec.Sort = "created desc"
		// TODO:
		ms, _, err = s.w.Content().ListArticle(ctx, &spec)
		if err != nil && err != ErrNoRows {
			logger().Infow("get Article", "error", err)
			return
		}
		if len(ms) == 0 {
			break
		}
		for i := range ms {
			err = s.upsertESArticle(ctx, &ms[i])
			if err != nil {
				return
			}
		}
		page++
	}
	return
}
func (s *contentStore) deleteESArticle(ctx context.Context, obj *cms1.Article) error {
	if obj == nil {
		return nil
	}
	err := DeleteESDoc(ctx, obj.IdentityTable(), obj.StringID())
	if err != nil {
		logger().Infow("DeleteESDoc", "index", obj.IdentityTable(), "error", err)
		return nil
	}
	return nil
}
