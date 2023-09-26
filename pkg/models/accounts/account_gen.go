// This file is generated - Do Not Edit.

package accounts

import (
	"fmt"

	comm "github.com/cupogo/andvari/models/comm"
	oid "github.com/cupogo/andvari/models/oid"
)

// 角色类型
type RoleType int8

const (
	RoleTypeNormal RoleType = 1 << iota //  1 普通用户
	RoleTypeAdmin                       //  2 管理员

	RoleTypeNone RoleType = 0 // none
)

func (z *RoleType) Decode(s string) error {
	switch s {
	case "0", "non", "none":
		*z = RoleTypeNone
	case "1", "nor", "normal", "Normal":
		*z = RoleTypeNormal
	case "2", "adm", "admin", "Admin":
		*z = RoleTypeAdmin
	default:
		return fmt.Errorf("invalid roleType: %q", s)
	}
	return nil
}
func (z RoleType) String() string {
	switch z {
	case RoleTypeNone:
		return "non"
	case RoleTypeNormal:
		return "nor"
	case RoleTypeAdmin:
		return "adm"
	default:
		return fmt.Sprintf("roleType %d", int8(z))
	}
}

// 账号状态
type AccountStatus int8

const (
	AccountStatusActive AccountStatus = 1 << iota //  1 active
	AccountStatusForbid                           //  2 forbid

	AccountStatusNone AccountStatus = 0 // none
)

func (z *AccountStatus) Decode(s string) error {
	switch s {
	case "0", "none":
		*z = AccountStatusNone
	case "1", "active", "Active":
		*z = AccountStatusActive
	case "2", "forbid", "Forbid":
		*z = AccountStatusForbid
	default:
		return fmt.Errorf("invalid accountStatus: %q", s)
	}
	return nil
}
func (z AccountStatus) String() string {
	switch z {
	case AccountStatusNone:
		return "none"
	case AccountStatusActive:
		return "active"
	case AccountStatusForbid:
		return "forbid"
	default:
		return fmt.Sprintf("accountStatus %d", int8(z))
	}
}

// consts of Account 账号
const (
	AccountTable = "auth_account"
	AccountAlias = "aa"
	AccountLabel = "account"
)

// Account 账号
type Account struct {
	comm.BaseModel `bun:"table:auth_account,alias:aa" json:"-"`

	comm.DefaultModel

	AccountBasic

	comm.MetaField
} // @name accountsAccount

type AccountBasic struct {
	// 登录名 唯一
	Username string `bun:"username,notnull,type:varchar(31),unique" extensions:"x-order=A" form:"username" json:"username" pg:"username,notnull,type:varchar(31),unique"`
	// 昵称
	Nickname string `bun:"nickname,notnull,type:varchar(45)" extensions:"x-order=B" form:"nickname" json:"nickname" pg:"nickname,notnull,use_zero,type:varchar(45)"`
	// 头像路径
	AvatarPath string `bun:"avatar,notnull,type:varchar(97)" extensions:"x-order=C" form:"avatar" json:"avatar,omitempty" pg:"avatar,notnull,use_zero,type:varchar(97)"`
	// 角色类型: 1=普通账号，2=管理员
	RoleType RoleType `bun:"role_type,notnull,type:smallint" extensions:"x-order=D" form:"rt" json:"rt" pg:"role_type,notnull,type:smallint,use_zero" swaggertype:"integer"`
	// 状态: 1=激活，2=禁用
	Status AccountStatus `bun:"status,notnull,type:smallint" extensions:"x-order=E" form:"status" json:"status" pg:"status,notnull,type:smallint,use_zero" swaggertype:"integer"`
	// 邮箱
	Email string `bun:"email,notnull,type:varchar(43)" extensions:"x-order=F" form:"email" json:"email,omitempty" pg:"email,notnull,use_zero,type:varchar(43)"`
	// 描述
	Description string `bun:",notnull" extensions:"x-order=G" form:"description" json:"description,omitempty" pg:",notnull,use_zero"`
	// 密码 (仅用于参数传递 Only used for parameter passing.)
	Password string `bun:"-" extensions:"x-order=H" form:"password" json:"password,omitempty" pg:"-"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" bun:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
} // @name accountsAccountBasic

type Accounts []Account

// Creating function call to it's inner fields defined hooks
func (z *Account) Creating() error {
	if z.IsZeroID() {
		z.SetID(oid.NewID(oid.OtAccount))
	}

	return z.DefaultModel.Creating()
}
func NewAccountWithBasic(in AccountBasic) *Account {
	obj := &Account{
		AccountBasic: in,
	}
	_ = obj.MetaUp(in.MetaDiff)
	return obj
}
func NewAccountWithID(id any) *Account {
	obj := new(Account)
	_ = obj.SetID(id)
	return obj
}
func (_ *Account) IdentityLabel() string {
	return AccountLabel
}
func (_ *Account) IdentityTable() string {
	return AccountTable
}
func (_ *Account) IdentityAlias() string {
	return AccountAlias
}

type AccountSet struct {
	// 登录名 唯一
	Username *string `extensions:"x-order=A" json:"username"`
	// 昵称
	Nickname *string `extensions:"x-order=B" json:"nickname"`
	// 头像路径
	AvatarPath *string `extensions:"x-order=C" form:"avatar" json:"avatar,omitempty"`
	// 角色类型: 1=普通账号，2=管理员
	RoleType *RoleType `extensions:"x-order=D" json:"rt" swaggertype:"integer"`
	// 状态: 1=激活，2=禁用
	Status *AccountStatus `extensions:"x-order=E" json:"status" swaggertype:"integer"`
	// 邮箱
	Email *string `extensions:"x-order=F" form:"email" json:"email,omitempty"`
	// 描述
	Description *string `extensions:"x-order=G" form:"description" json:"description,omitempty"`
	// 密码 (仅用于参数传递 Only used for parameter passing.)
	Password *string `extensions:"x-order=H" form:"password" json:"password,omitempty"`
	// for meta update
	MetaDiff *comm.MetaDiff `json:"metaUp,omitempty" swaggerignore:"true"`
} // @name accountsAccountSet

func (z *Account) SetWith(o AccountSet) {
	if o.Username != nil && z.Username != *o.Username {
		z.LogChangeValue("username", z.Username, o.Username)
		z.Username = *o.Username
	}
	if o.Nickname != nil && z.Nickname != *o.Nickname {
		z.LogChangeValue("nickname", z.Nickname, o.Nickname)
		z.Nickname = *o.Nickname
	}
	if o.AvatarPath != nil && z.AvatarPath != *o.AvatarPath {
		z.LogChangeValue("avatar", z.AvatarPath, o.AvatarPath)
		z.AvatarPath = *o.AvatarPath
	}
	if o.RoleType != nil && z.RoleType != *o.RoleType {
		z.LogChangeValue("role_type", z.RoleType, o.RoleType)
		z.RoleType = *o.RoleType
	}
	if o.Status != nil && z.Status != *o.Status {
		z.LogChangeValue("status", z.Status, o.Status)
		z.Status = *o.Status
	}
	if o.Email != nil && z.Email != *o.Email {
		z.LogChangeValue("email", z.Email, o.Email)
		z.Email = *o.Email
	}
	if o.Description != nil && z.Description != *o.Description {
		z.LogChangeValue("description", z.Description, o.Description)
		z.Description = *o.Description
	}
	if o.Password != nil && z.Password != *o.Password {
		z.Password = *o.Password
	}
	if o.MetaDiff != nil && z.MetaUp(o.MetaDiff) {
		z.SetChange("meta")
	}
}
func (in *AccountBasic) MetaAddKVs(args ...any) *AccountBasic {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}
func (in *AccountSet) MetaAddKVs(args ...any) *AccountSet {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}

// consts of AccountPasswd 账号密码
const (
	AccountPasswdTable = "auth_account_passwd"
	AccountPasswdAlias = "aap"
	AccountPasswdLabel = "accountPasswd"
)

// AccountPasswd 账号密码
type AccountPasswd struct {
	comm.BaseModel `bun:"table:auth_account_passwd,alias:aap" json:"-"`

	comm.DefaultModel

	AccountPasswdBasic

	comm.MetaField
} // @name accountsAccountPasswd

type AccountPasswdBasic struct {
	// 密码
	Password string `bun:",notnull,type:varchar(99)" extensions:"x-order=A" form:"password" json:"password" pg:",notnull,type:varchar(99)"`
	// for meta update
	MetaDiff *comm.MetaDiff `bson:"-" bun:"-" json:"metaUp,omitempty" pg:"-" swaggerignore:"true"`
} // @name accountsAccountPasswdBasic

type AccountPasswds []AccountPasswd

// Creating function call to it's inner fields defined hooks
func (z *AccountPasswd) Creating() error {
	if z.IsZeroID() {
		return comm.ErrEmptyID
	}

	return z.DefaultModel.Creating()
}
func NewAccountPasswdWithBasic(in AccountPasswdBasic) *AccountPasswd {
	obj := &AccountPasswd{
		AccountPasswdBasic: in,
	}
	_ = obj.MetaUp(in.MetaDiff)
	return obj
}
func NewAccountPasswdWithID(id any) *AccountPasswd {
	obj := new(AccountPasswd)
	_ = obj.SetID(id)
	return obj
}
func (_ *AccountPasswd) IdentityLabel() string {
	return AccountPasswdLabel
}
func (_ *AccountPasswd) IdentityTable() string {
	return AccountPasswdTable
}
func (_ *AccountPasswd) IdentityAlias() string {
	return AccountPasswdAlias
}

type AccountPasswdSet struct {
	// 密码
	Password *string `extensions:"x-order=A" json:"password"`
	// for meta update
	MetaDiff *comm.MetaDiff `json:"metaUp,omitempty" swaggerignore:"true"`
} // @name accountsAccountPasswdSet

func (z *AccountPasswd) SetWith(o AccountPasswdSet) {
	if o.Password != nil && z.Password != *o.Password {
		z.LogChangeValue("password", z.Password, o.Password)
		z.Password = *o.Password
	}
	if o.MetaDiff != nil && z.MetaUp(o.MetaDiff) {
		z.SetChange("meta")
	}
}
func (in *AccountPasswdBasic) MetaAddKVs(args ...any) *AccountPasswdBasic {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}
func (in *AccountPasswdSet) MetaAddKVs(args ...any) *AccountPasswdSet {
	in.MetaDiff = comm.MetaDiffAddKVs(in.MetaDiff, args...)
	return in
}
