package attrs

type ListAttr struct {
	items []interface{}
}

func NewListAttr() *ListAttr {
	return &ListAttr{
		items: []interface{}{},
	}
}
