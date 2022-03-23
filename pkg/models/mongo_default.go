package models

import (
	"time"
)

// IDField struct contain model's ID field.
type IDField struct {
	ID string `bson:"_id,omitempty" json:"id" redis:"id" extensions:"x-order=/"` // 主键
}

// DateFields struct contain `createdAt` and `updatedAt`
// fields that autofill on insert/update model.
type DateFields struct {
	CreatedAt time.Time `bson:"createdAt" json:"createdAt" redis:"createdAt" extensions:"x-order=k"` // 创建时间
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt" redis:"updatedAt" extensions:"x-order=l"` // 变更时间
}

// PrepareID method prepare id value to using it as id in filtering,...
// e.g convert hex-string id value to bson.ObjectId
func (f *IDField) PrepareID(id interface{}) (interface{}, error) {
	if idStr, ok := id.(string); ok {
		return idStr, nil
	}

	// Otherwise id must be ObjectId
	return id, nil
}

// GetID method return model's id
func (f *IDField) GetID() interface{} {
	return f.ID
}

// SetID set id value of model's id field.
func (f *IDField) SetID(id interface{}) {
	f.ID = id.(string)
}

//--------------------------------
// DateField methods
//--------------------------------

// Creating hook used here to set `created_at` field
// value on inserting new model into database.
func (f *DateFields) Creating() error {
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now()
	}

	return nil
}

// Saving hook used here to set `updated_at` field value
// on create/update model.
func (f *DateFields) Saving() error {
	f.UpdatedAt = time.Now()

	return nil
}

// DefaultModel struct contain model's default fields.
type DefaultModel struct {
	IDField    `bson:",inline"`
	DateFields `bson:",inline"`
	// 创建者ID
	CreatorID string `bson:"creatorID,omitempty" json:"creatorID,omitempty"  extensions:"x-order=m"`
}

// Creating function call to it's inner fields defined hooks
func (model *DefaultModel) Creating() error {
	return model.DateFields.Creating()
}

// Saving function call to it's inner fields defined hooks
func (model *DefaultModel) Saving() error {
	return model.DateFields.Saving()
}

// GetCreatorID ...
func (model *DefaultModel) GetCreatorID() string {
	return model.CreatorID
}

func (model *DefaultModel) SetCreatorID(id string) {
	model.CreatorID = id
}

type MWithIDer interface {
	GetID() interface{}
	SetID(id interface{})
	PrepareID(id interface{}) (interface{}, error)
}

// MWithCreator 基础模型
type MWithCreator interface {
	GetCreatorID() string
	SetCreatorID(id string)
}

// Sifter 查询过滤器
type Sifter interface {
	Sift() BD
}

// MDftSpec 默认的查询条件
type MDftSpec struct {
	// 主键编号（集）
	IDs []string `form:"id[]" json:"ids"  extensions:"x-order=2"`
	// 创建者ID
	CreatorID string `form:"creatorID" json:"creatorID"  extensions:"x-order=3"`
} // @name DefaultSpec

func (spec *MDftSpec) Sift() (qd BD) {
	qd = BD{}
	qd = SiftStrIn(qd, "_id", spec.IDs)
	qd = SiftStr(qd, "creatorID", spec.CreatorID)

	return
}

type StringsDiff struct {
	Newest  []string `json:"newest" validate:"dive"`  // 新增的字串集
	Removed []string `json:"removed" validate:"dive"` // 删除的字串集
} // @name StringsDiff

type TextSearchField struct {
	TextKeyword string `bson:"text_keyword" json:"-"` // 以空格区别的关键词
} // @name TextSearchField
