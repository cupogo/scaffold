package gens

// consts of files
const (
	headerComment = "This file is generated - Do Not Edit."

	storepkg = "stores"
	storewf  = "wrap.go"
	storewn  = "Wrap"
	storein  = "Storage"
)

// models
const (
	modelDefault = "DefaultModel"
	modelDunce   = "DunceModel"
	modelSerial  = "SerialModel"
)

// dbcode
type DbCode string

const (
	DbBun DbCode = "bun" // github.com/uptrace/bun
	DbPgx DbCode = "pgx" // github.com/go-pg/pg/v10
	DbMgm DbCode = "mgm" // go.mongodb.org, ORM with github.com/kamva/mgm/v3
)

const (
	relBelongsTo = "belongs-to"
	relHasOne    = "has-one"
	relMasMany   = "has-many"
)

// consts of qual
const (
	ginQual = "github.com/gin-gonic/gin"

	metaField       = "*.MetaField"
	ownerField      = "*.OwnerField"
	auditField      = "*.AuditFields"
	textSearchField = "*.TextSearchField"
)

// consts of created
const (
	createdField  = "CreatedAt"
	createdColumn = "created"
)

// consts of hooks
const (
	beforeSaving   = "beforeSaving"
	afterSaving    = "afterSaving"
	beforeCreating = "beforeCreating"
	afterCreating  = "afterCreating"
	beforeUpdating = "beforeUpdating"
	afterUpdating  = "afterUpdating"
	beforeDeleting = "beforeDeleting"
	afterDeleting  = "afterDeleting"
	afterCreated   = "afterCreated"
	afterUpdated   = "afterUpdated"
	afterDeleted   = "afterDeleted"
	afterLoad      = "afterLoad"
	afterList      = "afterList"
	beforeList     = "beforeList"
	upsertES       = "upsertES"
	deleteES       = "deleteES"
	errorLoad      = "errorLoad"
)

const (
	TagSwaggerIgnore = "swaggerignore"
	TagSwaggerType   = "swaggertype"
	TagExtensions    = "extensions"
)

type CompareType string

const (
	CompareScalar  CompareType = "scalar"
	CompareEqualTo CompareType = "equalTo"
)

const (
	MigrateES = "MigrateES"
)
