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
	ginQual = "github.com/gin-gonic/gin"
	oidQual = "hyyl.xyz/cupola/aurora/pkg/models/oid"

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
