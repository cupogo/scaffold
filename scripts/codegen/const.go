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

// consts of qual
const (
	ginQual = "github.com/gin-gonic/gin"
	oidQual = "github.com/cupogo/andvari/models/oid"

	utilsQual = "github.com/cupogo/andvari/utils"

	metaField       = "comm.MetaField"
	ownerField      = "comm.OwnerField"
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
	afterCreated   = "afterCreated"
	afterLoad      = "afterLoad"
	afterList      = "afterList"
)
