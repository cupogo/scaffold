package stores

import (
	"context"
	"sync"

	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/cupogo/andvari/database/embeds"
	"github.com/cupogo/andvari/models/comm"
	"github.com/cupogo/andvari/stores/pgx"
	"github.com/cupogo/andvari/utils"
	"github.com/cupogo/andvari/utils/zlog"

	"github.com/cupogo/scaffold/pkg/settings"
)

type ormDB = pgx.IDB //nolint
type ormQuery = pgx.SelectQuery
type pgDB = pgx.IDB      //nolint
type pgTx = pgx.Tx       //nolint
type pgIdent = pgx.Ident //nolint
type pgSafe = pgx.Safe   //nolint

type PageSpec = comm.PageSpec
type ModelSpec = pgx.ModelSpec
type TextSearchSpec = pgx.TextSearchSpec
type StringsDiff = pgx.StringsDiff

// vars
var (
	pgIn    = pgx.In          //nolint
	pgArray = pgdialect.Array //nolint

	ErrNoRows   = pgx.ErrNoRows
	ErrNotFound = pgx.ErrNotFound
	ErrEmptyKey = pgx.ErrEmptyKey

	dbGet              = pgx.Get
	dbFirst            = pgx.First
	dbLast             = pgx.Last
	queryOne           = pgx.QueryOne
	queryList          = pgx.QueryList
	queryPager         = pgx.QueryPager      //nolint
	getModelWherePK    = pgx.ModelWithPK     //nolint
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
	isZero    = utils.IsZero  //nolint

	ContextWithColumns  = pgx.ContextWithColumns
	ColumnsFromContext  = pgx.ColumnsFromContext
	ContextWithRelation = pgx.ContextWithRelation
	RelationFromContext = pgx.RelationFromContext

	RegisterModel = pgx.RegisterModel
)

func logger() zlog.Logger {
	return zlog.Get()
}

func init() {
	pgx.RegisterDbFs(embeds.DBFS())
}

// vars ...
var (
	_ Storage = (*Wrap)(nil)

	dbOnce sync.Once
	dbX    *pgx.DB

	stoOnce sync.Once
	stoW    *Wrap
)

// Wrap implements Storages
type Wrap struct {
	db *pgx.DB

	contentStore *contentStore // gened
}

// NewWithDB return new instance of Wrap
func NewWithDB(db *pgx.DB) *Wrap {
	w := &Wrap{db: db}

	w.contentStore = newContentStore(w) // gened

	// more member stores
	return w
}

// SgtDB start and return a singleton instance of DB
// **Attention**: args only used with fist call
func SgtDB(args ...string) *pgx.DB {
	dbOnce.Do(func() {
		dsn := settings.Current.PgStoreDSN
		tscfg := settings.Current.PgTSConfig
		if len(args) > 0 && len(args[0]) > 0 {
			dsn = args[0]
			if len(args) > 1 {
				tscfg = args[1]
			}
		}
		var err error
		dbX, err = pgx.Open(dsn, tscfg, settings.Current.PgQueryDebug)
		if err != nil {
			logger().Panicw("connect to database fail", "err", err)
		}
	})
	return dbX
}

// Sgt start and return a singleton instance of Storage
func Sgt() *Wrap {
	stoOnce.Do(func() {
		stoW = NewWithDB(SgtDB())
	})
	return stoW
}

func (w *Wrap) Close() {
	_ = w.db.Close()
}

// dbOpModelMeta prepare meta from Context
func dbOpModelMeta(ctx context.Context, db ormDB, obj comm.ModelMeta, ups ...*comm.MetaDiff) {
	if len(ups) > 0 {
		_ = obj.MetaUp(ups[0])
	}
}
func (w *Wrap) Content() ContentStore { return w.contentStore } // Content gened
