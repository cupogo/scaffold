// This file is generated - Do Not Edit.

package stores

import (
	"context"

	"github.com/cupogo/scaffold/pkg/models/accs"
)

// type Account = accs.Account
// type AccountBasic = accs.AccountBasic
// type AccountPasswd = accs.AccountPasswd
// type AccountPasswdBasic = accs.AccountPasswdBasic
// type AccountPasswdSet = accs.AccountPasswdSet
// type AccountPasswds = accs.AccountPasswds
// type AccountSet = accs.AccountSet
// type AccountStatus = accs.AccountStatus
// type Accounts = accs.Accounts

func init() {
	RegisterModel((*accs.Account)(nil), (*accs.AccountPasswd)(nil))
}

type AccountStore interface {
	AccountStoreX

	ListAccount(ctx context.Context, spec *AccountSpec) (data accs.Accounts, total int, err error)
	GetAccount(ctx context.Context, id string) (obj *accs.Account, err error)
	CreateAccount(ctx context.Context, in accs.AccountBasic) (obj *accs.Account, err error)
	UpdateAccount(ctx context.Context, id string, in accs.AccountSet) error
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
	Status accs.AccountStatus `extensions:"x-order=C" form:"status" json:"status" swaggertype:"integer"`
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

func (s *accountStore) ListAccount(ctx context.Context, spec *AccountSpec) (data accs.Accounts, total int, err error) {
	total, err = s.w.db.ListModel(ctx, spec, &data)
	return
}
func (s *accountStore) GetAccount(ctx context.Context, id string) (obj *accs.Account, err error) {
	obj = new(accs.Account)
	if err = dbGet(ctx, s.w.db, obj, "username ILIKE ?", id); err != nil {
		err = s.w.db.GetModel(ctx, obj, id)
	}
	if err == nil {
		err = s.afterLoadAccount(ctx, obj)
	}
	return
}
func (s *accountStore) CreateAccount(ctx context.Context, in accs.AccountBasic) (obj *accs.Account, err error) {
	err = s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		obj = accs.NewAccountWithBasic(in)
		if obj.Username == "" {
			err = ErrEmptyKey
			return
		}
		if err = dbBeforeCreateAccount(ctx, tx, obj); err != nil {
			return
		}
		dbOpModelMeta(ctx, tx, obj)
		err = dbInsert(ctx, tx, obj, "username")
		if err == nil {
			err = dbAfterSaveAccount(ctx, tx, obj)
		}
		return err
	})
	return
}
func (s *accountStore) UpdateAccount(ctx context.Context, id string, in accs.AccountSet) error {
	exist := new(accs.Account)
	if err := getModelWithPKID(ctx, s.w.db, exist, id); err != nil {
		return err
	}
	exist.SetWith(in)
	return s.w.db.RunInTx(ctx, nil, func(ctx context.Context, tx pgTx) (err error) {
		exist.SetIsUpdate(true)
		if err = dbBeforeUpdateAccount(ctx, tx, exist); err != nil {
			return
		}
		dbOpModelMeta(ctx, tx, exist)
		if err = dbUpdate(ctx, tx, exist); err == nil {
			return dbAfterSaveAccount(ctx, tx, exist)
		}
		return
	})
}
func (s *accountStore) DeleteAccount(ctx context.Context, id string) error {
	obj := new(accs.Account)
	return s.w.db.DeleteModel(ctx, obj, id)
}
