package gens

import (
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
)

var (
	swdb  jen.Code = jen.Id("s").Dot("w").Dot("db")
	jactx jen.Code = jen.Id("ctx").Id("context.Context")
	jadbO jen.Code = jen.Id("db").Id("ormDB")
	jrctx jen.Code = jen.Id("c").Dot("Request").Dot("Context").Call()

	methodsMongo = map[rune]string{
		'L': "mgList",
		'C': "mgCreate",
		'U': "mgUpdate",
		'D': "s.w.deleteModel",
		'G': "mgGetWithID",
	}
	methodsPGx = map[rune]string{
		'L': "s.w.db.ListModel",
		'C': "dbInsert",
		'U': "dbUpdate",
		'D': "s.w.db.DeleteModel",
		'G': "dbGetWithPKID",
	}

	commonAbbrs = []string{
		"HR",
		"ID",
	}
)

func IsCommonAbbr(s string) (string, bool) {
	k := strings.ToUpper(s)
	if slices.Contains(commonAbbrs, k) {
		return k, true
	}
	return s, false
}
