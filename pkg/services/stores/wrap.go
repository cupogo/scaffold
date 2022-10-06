package stores

import (
	"context"

	redis "github.com/go-redis/redis/v8"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/cupogo/andvari/database/embeds"
	"github.com/cupogo/andvari/models/comm"
	"github.com/cupogo/andvari/stores/pgx"
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

	ErrNoRows = pgx.ErrNoRows

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

// Wrap implements Storages
type Wrap struct {
	db *pgx.DB
	rc redis.UniversalClient

	contentStore *contentStore // gened
}

// NewWithDB ...
func NewWithDB(db *pgx.DB, rc redis.UniversalClient) *Wrap {
	w := &Wrap{db: db, rc: rc}

	w.contentStore = &contentStore{w: w} // gened

	// more member stores
	return w
}

// New with dsn, db, redis, only once
func New(args ...string) (*Wrap, error) {
	db, rc, err := OpenBases(args...)
	if err != nil {
		return nil, err
	}
	return NewWithDB(db, rc), nil
}

// OpenBases open multiable databases
func OpenBases(args ...string) (db *pgx.DB, rc redis.UniversalClient, err error) {
	dsn := settings.Current.PgStoreDSN
	if len(args) > 0 && len(args[0]) > 0 {
		dsn = args[0]
	}
	db, err = pgx.Open(dsn, settings.Current.PgTSConfig, settings.Current.PgQueryDebug)
	if err != nil {
		return
	}

	redisURI := settings.Current.RedisURI
	opt, err := redis.ParseURL(redisURI)
	if err != nil {
		logger().Warnw("prase redisURI fail", "uri", redisURI, "err", err)
		return
	}
	rc = redis.NewClient(opt)

	return
}

// opModelMeta prepare values from Context
func (w *Wrap) opModelMeta(ctx context.Context, obj comm.ModelMeta, ups ...*comm.MetaDiff) {

	if mm, ok := obj.(interface{ MetaUp(*comm.MetaDiff) bool }); ok && len(ups) > 0 {
		_ = mm.MetaUp(ups[0])
	}
}
func (w *Wrap) Content() ContentStore { return w.contentStore } // Content gened
