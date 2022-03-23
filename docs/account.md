# 账号

## Models

### `Account` 账号信息

```go

// RoleType 角色类型
type RoleType int32

// AccountRole
const (
	roleTypeNone RoleType = iota // 0
	RoleTypeNormal               // 1
	RoleTypeAdmin                // 2
)


```

Name|BSON|JSON|Type|Doc
:---|:---|:---|:---|:--
`DefaultModel`|`,inline`|||
`Name`|`name`|`name`|`string`|名称
`RoleType`|`role_type`|`rt`|`RoleType`|角色类型


```go

type Accounts []Account

// IsAdmin return true if is admin type
func (a *Account) IsAdmin() bool {
	return a.RoleType == RoleTypeAdmin
}

```
