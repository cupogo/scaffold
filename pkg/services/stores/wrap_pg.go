//lint:file-ignore U1000 ignore unused code
package stores

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"

	"hyyl.xyz/cupola/andvari/models/comm"
	"hyyl.xyz/cupola/andvari/models/oid"
	"hyyl.xyz/cupola/andvari/stores/pgx"
)

type ormDB = orm.DB //nolint
type ormQuery = orm.Query
type pgDB = pg.DB //nolint
type pgTx = pg.Tx //nolint
type pgIdent = pg.Ident
type pgSafe = pg.Safe //nolint

type MDftSpec = pgx.ModelSpec // deprecated
type ModelSpec = pgx.ModelSpec
type TextSearchSpec = pgx.TextSearchSpec

// vars
var (
	pgIn      = pg.In      //nolint
	pgInMulti = pg.InMulti //nolint
	pgArray   = pg.Array   //nolint
	pgScan    = pg.Scan    //nolint

	ErrNoRows = pg.ErrNoRows

	queryPager         = pgx.QueryPager
	getModelWherePK    = pgx.ModelWherePK    //nolint
	getModelWithPKID   = pgx.ModelWithPKID   //nolint
	getModelWithPKOID  = pgx.ModelWithPKID   //nolint
	getModelWithUnique = pgx.ModelWithUnique //nolint
	dbInsert           = pgx.DoInsert
	dbUpdate           = pgx.DoUpdate
	dbDeleteT          = pgx.DoDeleteT     //nolint
	dbStoreSimple      = pgx.StoreSimple   //nolint
	dbStoreWithCall    = pgx.StoreWithCall //nolint

	sift      = pgx.Sift      //nolint
	siftEquel = pgx.SiftEquel //nolint
	siftILike = pgx.SiftICE   //nolint
	siftMatch = pgx.SiftMatch //nolint
	siftGreat = pgx.SiftGreat //nolint
	siftLess  = pgx.SiftLess  //nolint
	siftOID   = pgx.SiftOID   //nolint
	siftOIDs  = pgx.SiftOIDs  //nolint
	siftDate  = pgx.SiftDate  //nolint
)

var (
	alltables []any
)

type OID = oid.OID

// opModelMeta prepare values from Context
func (w *Wrap) opModelMeta(ctx context.Context, obj comm.ModelCreator, ups ...*comm.MetaDiff) {

	if mm, ok := obj.(interface{ MetaUp(*comm.MetaDiff) bool }); ok && len(ups) > 0 {
		_ = mm.MetaUp(ups[0])
	}
}
