package stores

import (
	"context"

	"github.com/cupogo/scaffold/pkg/models/accs"
)

type AccountStoreX interface {
}

func dbBeforeUpdateAccount(ctx context.Context, db ormDB, obj *accs.Account) error {
	// TODO:
	return nil
}
func dbBeforeCreateAccount(ctx context.Context, db ormDB, obj *accs.Account) error {
	// TODO:
	return nil
}
func dbAfterSaveAccount(ctx context.Context, db ormDB, obj *accs.Account) error {
	// TODO:
	return nil
}
func (s *accountStore) afterLoadAccount(ctx context.Context, obj *accs.Account) error {
	// TODO: need implement
	return nil
}
