package gens

import "github.com/dave/jennifer/jen"

var (
	swdb  jen.Code = jen.Id("s").Dot("w").Dot("db")
	jactx jen.Code = jen.Id("ctx").Id("context.Context")
	jadbO jen.Code = jen.Id("db").Id("ormDB")
	jrctx jen.Code = jen.Id("c").Dot("Request").Dot("Context").Call()
)
