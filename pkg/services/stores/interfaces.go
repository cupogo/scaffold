package stores

import (
	"context"
)

type ModelIdentity interface {
	IdentityLabel() string
	IdentityTable() string
	IdentityAlias() string
	StringID() string
	MetaSet(key string, value any)
}

type Storage interface {
	Content() ContentStore // gened
	Account() AccountStore // gened
}

var UpsertESDoc func(ctx context.Context, index string, mi ModelIdentity) error
var DeleteESDoc func(ctx context.Context, index, id string) error

func RegisterESMigrate(ModelIdentity, func(context.Context) error) {

}
