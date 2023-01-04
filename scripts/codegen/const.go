package main

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
	createdName   = "CreatedAt"
	createdColumn = "created"
)

// consts of hooks
const (
	beforeSaving   = "beforeSaving"
	afterSaving    = "afterSaving"
	beforeCreating = "beforeCreating"
	beforeUpdating = "beforeUpdating"
	beforeDeleting = "beforeDeleting"
	afterDeleting  = "afterDeleting"
	afterCreated   = "afterCreated"
	afterUpdated   = "afterUpdated"
	afterDeleted   = "afterDeleted"
	afterLoad      = "afterLoad"
	afterList      = "afterList"
	beforeList     = "beforeList"
)

const (
	TagSwaggerIgnore = "swaggerignore"
	TagExtensions    = "extensions"
)

type CompareType string

const (
	CompareScalar  CompareType = "scalar"
	CompareEqualTo CompareType = "equalTo"
)
