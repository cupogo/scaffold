package stores

import (
	"context"

	"github.com/cupogo/scaffold/pkg/models/accounts"
)

type AccountStoreX interface {
}

func dbBeforeUpdateAccount(ctx context.Context, db ormDB, obj *accounts.Account) error {
	// TODO:
	return nil
}
func dbBeforeCreateAccount(ctx context.Context, db ormDB, obj *accounts.Account) error {
	// TODO:
	return nil
}
func dbAfterSaveAccount(ctx context.Context, db ormDB, obj *accounts.Account) error {
	// TODO:
	return nil
}
func (s *accountStore) afterLoadAccount(ctx context.Context, obj *accounts.Account) error {
	// TODO: need implement
	return nil
}
