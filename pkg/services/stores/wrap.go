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
// nolint
var (
	pgIn    = pgx.In
	pgArray = pgdialect.Array

	ErrNoRows   = pgx.ErrNoRows
	ErrNotFound = pgx.ErrNotFound
	ErrEmptyKey = pgx.ErrEmptyKey

	dbGet              = pgx.Get
	dbFirst            = pgx.First
	dbLast             = pgx.Last
	queryOne           = pgx.QueryOne
	queryList          = pgx.QueryList
	queryPager         = pgx.QueryPager
	getModelWherePK    = pgx.ModelWithPK
	getModelWithPKID   = pgx.ModelWithPKID
	getModelWithPKOID  = pgx.ModelWithPKID
	getModelWithUnique = pgx.ModelWithUnique
	dbInsert           = pgx.DoInsert
	dbUpdate           = pgx.DoUpdate
	dbDeleteT          = pgx.DoDeleteT
	dbStoreSimple      = pgx.StoreSimple
	dbMetaUp           = pgx.DoMetaUp
	dbOpModelMeta      = pgx.DoMetaUp

	sift      = pgx.Sift
	siftEquel = pgx.SiftEquel
	siftICE   = pgx.SiftICE
	siftMatch = pgx.SiftMatch
	siftOID   = pgx.SiftOID
	siftOIDs  = pgx.SiftOIDs
	siftDate  = pgx.SiftDate
	isZero    = utils.IsZero

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
	pgx.RegisterMetaUp(dbModelMetaUps)
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

// dbModelMetaUps all local metaUps
func dbModelMetaUps(ctx context.Context, db ormDB, obj pgx.Model) {
	// more
}
func (w *Wrap) Content() ContentStore { return w.contentStore } // Content gened
