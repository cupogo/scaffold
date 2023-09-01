// This file is generated - Do Not Edit.

package stores

import (
	"context"

	"github.com/cupogo/scaffold/pkg/models/accounts"
)

// type Account = accounts.Account
// type AccountBasic = accounts.AccountBasic
// type AccountPasswd = accounts.AccountPasswd
// type AccountPasswdBasic = accounts.AccountPasswdBasic
// type AccountPasswdSet = accounts.AccountPasswdSet
// type AccountPasswds = accounts.AccountPasswds
// type AccountSet = accounts.AccountSet
// type AccountStatus = accounts.AccountStatus
// type Accounts = accounts.Accounts

func init() {
	RegisterModel((*accounts.Account)(nil), (*accounts.AccountPasswd)(nil))
}

type AccountStore interface {
	AccountStoreX

	ListAccount(ctx context.Context, spec *AccountSpec) (data accounts.Accounts, total int, err error)
	GetAccount(ctx context.Context, id string) (obj *accounts.Account, err error)
	CreateAccount(ctx context.Context, in accounts.AccountBasic) (obj *accounts.Account, err error)
	UpdateAccount(ctx context.Context, id string, in accounts.AccountSet) error
	DeleteAccount(ctx context.Context, id string) error
}

type AccountSpec struct {
	PageSpec
	ModelSpec

	// 登录名 唯一
	Username string `extensions:"x-order=A" form:"username" json:"username"`
	// 昵称
	Nickname string `extensions:"x-order=B" form:"nickname" json:"nickname"`
	// 状态: 1=激活，2=禁用
	Status accounts.AccountStatus `extensions:"x-order=C" form:"status" json:"status" swaggertype:"integer"`
	// 邮箱
	Email string `extensions:"x-order=D" form:"email" json:"email,omitempty"`
	// 全部字段
	WithAll bool `extensions:"x-order=E" form:"all" json:"all"`
}

func (spec *AccountSpec) Sift(q *ormQuery) *ormQuery {
	q = spec.ModelSpec.Sift(q)
	q, _ = siftMatch(q, "username", spec.Username, false)
	q, _ = siftMatch(q, "nickname", spec.Nickname, false)
	q, _ = siftEqual(q, "status", spec.Status, false)
	q, _ = siftMatch(q, "email", spec.Email, false)

	return q
}

type accountStore struct {
	w *Wrap
}

func (s *accountStore) ListAccount(ctx context.Context, spec *AccountSpec) (data accounts.Accounts, total int, err error) {
	total, err = s.w.db.ListModel(ctx, spec, &data)
	return
}
func (s *accountStore) GetAccount(ctx context.Context, id string) (obj *accounts.Account, err error) {
	obj, err = GetAccount(ctx, s.w.db, id, ColumnsFromContext(ctx)...)
	if err == nil {
		err = s.afterLoadAccount(ctx, obj)
	}
	return
}
func (s *accountStore) CreateAccount(ctx context.Context, in accounts.AccountBasic) (obj *accounts.Account, err error) {
	err = s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		obj = accounts.NewAccountWithBasic(in)
		if obj.Username == "" {
			err = ErrEmptyKey
			return
		}
		if err = dbBeforeCreateAccount(ctx, tx, obj); err != nil {
			return
		}
		dbMetaUp(ctx, tx, obj)
		err = dbInsert(ctx, tx, obj, "username")
		if err == nil {
			err = dbAfterSaveAccount(ctx, tx, obj)
		}
		return err
	})
	return
}
func (s *accountStore) UpdateAccount(ctx context.Context, id string, in accounts.AccountSet) error {
	return s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		exist := new(accounts.Account)
		if err = dbGetWithPKID(ctx, tx, exist, id); err != nil {
			return err
		}
		exist.SetWith(in)
		exist.SetIsUpdate(true)
		if err = dbBeforeUpdateAccount(ctx, tx, exist); err != nil {
			return err
		}
		dbMetaUp(ctx, tx, exist)
		if err = dbUpdate(ctx, tx, exist); err != nil {
			return err
		}
		return dbAfterSaveAccount(ctx, tx, exist)
	})
}
func (s *accountStore) DeleteAccount(ctx context.Context, id string) error {
	obj := new(accounts.Account)
	return s.w.db.DeleteModel(ctx, obj, id)
}

func GetAccount(ctx context.Context, db ormDB, id string, cols ...string) (obj *accounts.Account, err error) {
	obj = new(accounts.Account)
	if err = dbGetWith(ctx, db, obj, "username", "ILIKE", id, cols...); err != nil {
		err = dbGetWithPKID(ctx, db, obj, id, cols...)
	}
	return
}
