//lint:file-ignore U1000 ignore unused code
package stores

import (
	"context"

	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/cupogo/andvari/models/comm"
	"github.com/cupogo/andvari/stores/pgx"
)

type ormDB = pgx.IDB //nolint
type ormQuery = pgx.SelectQuery
type pgDB = pgx.IDB //nolint
type pgTx = pgx.Tx  //nolint
type pgIdent = pgx.Ident
type pgSafe = pgx.Safe //nolint

type ModelSpec = pgx.ModelSpec
type TextSearchSpec = pgx.TextSearchSpec

// vars
var (
	pgIn    = pgx.In          //nolint
	pgArray = pgdialect.Array //nolint

	ErrNoRows = pgx.ErrNoRows

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
	siftICE   = pgx.SiftICE   //nolint
	siftMatch = pgx.SiftMatch //nolint
	siftOID   = pgx.SiftOID   //nolint
	siftOIDs  = pgx.SiftOIDs  //nolint
	siftDate  = pgx.SiftDate  //nolint

	ContextWithColumns  = pgx.ContextWithColumns
	ColumnsFromContext  = pgx.ColumnsFromContext
	ContextWithRelation = pgx.ContextWithRelation
	RelationFromContext = pgx.RelationFromContext
)

var (
	alltables []any
)

// opModelMeta prepare values from Context
func (w *Wrap) opModelMeta(ctx context.Context, obj comm.ModelMeta, ups ...*comm.MetaDiff) {

	if mm, ok := obj.(interface{ MetaUp(*comm.MetaDiff) bool }); ok && len(ups) > 0 {
		_ = mm.MetaUp(ups[0])
	}
}
