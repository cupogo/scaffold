package stores

import (
	"context"

	"github.com/cupogo/scaffold/pkg/models/cms1"
)

func dbBeforeSaveArticle(ctx context.Context, db ormDB, obj *cms1.Article) error {
	return nil
}

func dbAfterDeleteArticle(ctx context.Context, db ormDB, obj *cms1.Article) error {
	return nil
}
