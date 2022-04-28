package stores

import (
	redis "github.com/go-redis/redis/v8"

	"hyyl.xyz/cupola/aurora/pkg/services/utils/pgx"
	"hyyl.xyz/cupola/aurora/pkg/settings"
)

// Wrap implements Storages
type Wrap struct {
	db *pgx.DB
	rc *redis.Client
}

// NewWithDB ...
func NewWithDB(db *pgx.DB, rc *redis.Client) *Wrap {
	w := &Wrap{db: db, rc: rc}

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
func OpenBases(args ...string) (db *pgx.DB, rc *redis.Client, err error) {
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
