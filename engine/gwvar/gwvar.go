package gwvar

import "expvar"

type Bool struct {
	val *expvar.Int
}

func NewBool(name string) *Bool {
	return &Bool{
		val: expvar.NewInt("name"),
	}
}

func (b *Bool) Value() bool {
	return b.val.Value() > 0
}

func (b *Bool) Set(v bool) {
	if v {
		b.val.Set(1)
	} else {
		b.val.Set(0)
	}
}

var (
	IsDeploymentReady = NewBool("IsDeploymentReady")
)
