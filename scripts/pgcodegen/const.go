package main

// consts of files
const (
	headerComment = "This file is generated - Do Not Edit."

	storepkg = "stores"
	storewf  = "wrap.go"
	storewn  = "Wrap"
	storein  = "Storage"
)

// consts of qual
const (
	oidQual  = "hyyl.xyz/cupola/aurora/pkg/models/oid"
	errsQual = "hyyl.xyz/cupola/aurora/pkg/services/errors"
	utilQual = "hyyl.xyz/cupola/aurora/pkg/services/utils"

	metaField       = "comm.MetaField"
	auditField      = "evnt.AuditFields"
	textSearchField = "comm.TextSearchField"
)

// consts of hooks
const (
	beforeSaving   = "beforeSaving"
	afterSaving    = "afterSaving"
	beforeCreating = "beforeCreating"
	beforeUpdating = "beforeUpdating"
	beforeDeleting = "beforeDeleting"
	afterDeleting  = "afterDeleting"
	afterLoad      = "afterLoad"
	afterCreated   = "afterCreated"
)

// consts of api
const (
	ginQual  = "github.com/gin-gonic/gin"
	respQual = "hyyl.xyz/cupola/aurora/pkg/web/resp"
)
