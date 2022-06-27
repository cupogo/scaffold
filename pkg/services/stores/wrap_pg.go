//lint:file-ignore U1000 ignore unused code
package stores

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"

	"hyyl.xyz/cupola/aurora/pkg/models"
	"hyyl.xyz/cupola/aurora/pkg/models/comm"
	"hyyl.xyz/cupola/aurora/pkg/models/oid"
	"hyyl.xyz/cupola/aurora/pkg/services/utils/pgx"
)

type ormDB = orm.DB //nolint
type ormQuery = orm.Query
type pgDB = pg.DB //nolint
type pgTx = pg.Tx //nolint
type pgIdent = pg.Ident
type pgSafe = pg.Safe //nolint
type MDftSpec = pgx.MDftSpec

// vars
var (
	pgIn      = pg.In      //nolint
	pgInMulti = pg.InMulti //nolint
	pgArray   = pg.Array   //nolint
	pgScan    = pg.Scan    //nolint

	ErrNoRows = pg.ErrNoRows

	queryPager         = pgx.QueryPager
	getModelWherePK    = pgx.ModelWherePK //nolint
	getModelWithPKOID  = pgx.ModelWithPKOID
	getModelWithUnique = pgx.ModelWithUnique //nolint
	dbInsert           = pgx.DoInsert
	dbUpdate           = pgx.DoUpdate
	dbStoreSimple      = pgx.StoreSimple   //nolint
	dbStoreWithCall    = pgx.StoreWithCall //nolint

	sift      = pgx.Sift      //nolint
	siftEquel = pgx.SiftEquel //nolint
	siftILike = pgx.SiftILike //nolint
	siftGreat = pgx.SiftGreat //nolint
	siftLess  = pgx.SiftLess  //nolint
	siftOID   = pgx.SiftOID   //nolint
)

var (
	alltables []any
)

//nolint
type applier func(query *orm.Query) (*orm.Query, error)

type Model = models.Model
type OID = oid.OID

// opModelMeta prepare values from Context
func (w *Wrap) opModelMeta(ctx context.Context, obj models.ModelCreator, ups ...*comm.MetaUp) {

	if mm, ok := obj.(interface{ UpMeta(*comm.MetaUp) bool }); ok && len(ups) > 0 {
		_ = mm.UpMeta(ups[0])
	}
}
