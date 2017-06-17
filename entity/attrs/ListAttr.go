package attrs

const (
	CL_SET        = iota
	CL_GET        = iota
	CL_APPEND     = iota
	CL_POP        = iota
	CL_APPENDLEFT = iota
	CL_POPLEFT    = iota
)

type changelogRecord struct {
	kind  int
	index int
}

type ListAttr struct {
	parent    interface{}
	items     []interface{}
	changelog []changelogRecord
}

func NewListAttr(parent interface{}) *ListAttr {
	return &ListAttr{
		parent:    parent,
		items:     []interface{}{},
		changelog: []changelogRecord{},
	}
}

func (la *ListAttr) Append(val interface{}) {
	la.items = append(la.items, val)
	la.changelog = append(la.changelog, changelogRecord{CL_APPEND, 0})
}

func (la *ListAttr) Pop() (val interface{}) {
	l := len(la.items)
	val = la.items[l-1]
	la.items = la.items[:l-1]
	la.changelog = append(la.changelog, changelogRecord{CL_POP, 0})
	return
}

func (la *ListAttr) AppendLeft(val interface{}) {
	newitems := make([]interface{}, len(la.items)+1)
	copy(newitems[1:], la.items)
	newitems[0] = val
	la.items = newitems
	la.changelog = append(la.changelog, changelogRecord{CL_APPENDLEFT, 0})
}

func (la *ListAttr) PopLeft() (val interface{}) {
	val = la.items[0]
	la.items = la.items[1:]
	la.changelog = append(la.changelog, changelogRecord{CL_POPLEFT, 0})
	return
}

func (la *ListAttr) Set(index int, val interface{}) {
	la.items[index] = val
	la.changelog = append(la.changelog, changelogRecord{CL_SET, 0})
}

func (la *ListAttr) Get(index int) interface{} {
	res := la.items[index]
	la.changelog = append(la.changelog, changelogRecord{CL_GET, 0})
	return res
}
