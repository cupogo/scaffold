package resp

type ViewPatcher interface {
	PatchView()
}

// ResultData 特定数据集(带JSON数组和总数)，一般用在分页查询结果
type ResultData struct {
	Data  any `json:"data,omitempty"`  // 数据集数组
	Total int `json:"total,omitempty"` // 符合条件的总记录数
} // @name ResultData

func (dr *ResultData) PatchView() {
	if v, ok := dr.Data.(ViewPatcher); ok {
		v.PatchView()
	}
}

type ResultID struct {
	ID any `json:"id"` // 主键值，多数时候是字串
} // @name ResultID
