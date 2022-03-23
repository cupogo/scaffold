package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BD D represents a BSON Document.
type BD = bson.D // @name BSON_Document

// BE E represents a BSON element for a D.
type BE = bson.E // @name BSON_Element

// BM M is an unordered, concise representation of a BSON Document.
type BM = bson.M // @name BSON_Map

// BA An A represents a BSON array.
type BA = bson.A

// DateTime represents the BSON datetime value.
type DateTime = primitive.DateTime

var NewDateTimeFromTime = primitive.NewDateTimeFromTime

// MergeBM merge a map to other
func MergeBM(m, o BM) BM {
	if m == nil {
		m = BM{}
	}
	if o == nil {
		return m
	}
	for k, v := range o {
		m[k] = v
	}
	return m
}

// OpInStr ...
type OpInStr struct {
	In []string `bson:"$in,omitempty" json:"$in,omitempty"`
}

// OpInInt ...
type OpInInt struct {
	In []int64 `bson:"$in,omitempty" json:"$in,omitempty"`
}

// OpRegex ...
type OpRegex struct {
	Regex   string `bson:"$regex" json:"$regex"`
	Options string `bson:"$options" json:"$options"`
}

func Sift(qd BD, key string, val interface{}) BD {
	if val == nil || val == 0 || val == "" || val == false {
		return qd
	}
	return append(qd, BE{Key: key, Value: val})
}

func SiftStr(qd BD, key, val string) BD {
	if len(val) > 0 {
		return append(qd, BE{Key: key, Value: val})
	}
	return qd
}

func SiftStrIn(qd BD, key string, vals []string) BD {
	if len(vals) > 0 {
		return append(qd, BE{Key: key, Value: BM{"$in": vals}})
	}
	return qd
}

func SiftBool(qd BD, key string, val bool) BD {
	if val {
		return append(qd, BE{Key: key, Value: true})
	}
	{
		return append(qd, BE{Key: "$or", Value: BA{
			BD{{Key: key, Value: BD{{Key: "$exists", Value: false}}}},
			BD{{Key: key, Value: false}},
		}})
	}

}
