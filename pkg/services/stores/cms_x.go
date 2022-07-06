package stores

import (
	"context"

	"hyyl.xyz/cupola/scaffold/pkg/models/cms1"
)

func dbAfterDeleteArticle(ctx context.Context, db ormDB, obj *cms1.Article) error {
	return nil
}
