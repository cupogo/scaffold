package core

// EnumItem 枚举的条目
type EnumItem struct {
	ID     int    `extensions:"x-order=A" json:"id,omitempty" `
	Code   string `extensions:"x-order=B" json:"code"`
	Name   string `extensions:"x-order=C" json:"name"`
	Descr  string `extensions:"x-order=D" json:"descr,omitempty"`
	Parent string `extensions:"x-order=E" json:"parent,omitempty"`
} // @name coreEnumItem

type EnumItems []EnumItem

func (z EnumItem) EqualTo(o EnumItem) bool {
	return z.ID == o.ID && z.Code == o.Code && z.Name == o.Name &&
		z.Descr == o.Descr && z.Parent == o.Parent
}
func (z EnumItems) EqualTo(o EnumItems) bool {
	if len(z) != len(o) {
		return false
	}
	for i := 0; i < len(z); i++ {
		if !z[i].EqualTo(o[i]) {
			return false
		}
	}
	return true
}
