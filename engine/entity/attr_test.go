package entity

import (
	"math"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func TestAttrVals(t *testing.T) {
	v := uniformAttrType(float32(1.0))
	t.Logf("uniformAttrType %v %T", v, v)
	v = uniformAttrType(int32(1))
	t.Logf("uniformAttrType %v %T", v, v)
}

func TestMapAttr(t *testing.T) {
	m := NewMapAttr()
	m.AssignMap(map[string]interface{}{
		"int":     int(1),
		"int32":   int32(32),
		"int64":   int64(64),
		"float32": float32(32.0),
		"float64": float64(64.0),
		"bool":    true,
		"string":  "xxx",
		"bson.M": bson.M{
			"a": 1, "b": 2,
			"innerM1": bson.M{
				"a": 1, "b": 2,
			},
		},
		"subM": map[string]interface{}{
			"a": 1, "b": 2,
			"innerM1": map[string]interface{}{
				"a": 1, "b": 2,
			},
		},
		"list": []interface{}{1, 2, 3, map[string]interface{}{
			"a": 1, "b": 2,
			"innerM1": map[string]interface{}{
				"a": 1, "b": 2, "l": []interface{}{1, "xxx", false},
			},
		}},
	})

	t.Logf("AssignMap: %s", m)

	if !m.HasKey("int") {
		t.Fatalf("should has key")
	}

	if m.HasKey("not exist key") {
		t.Fatalf("should not has key")
	}

	if m.GetInt("int") != 1 {
		t.Fatalf("wrong value")
	}
	if m.GetInt("int32") != 32 {
		t.Fatalf("wrong value")
	}
	if m.GetInt("int64") != 64 {
		t.Fatalf("wrong value")
	}
	if m.GetInt("not exist key") != 0 {
		t.Fatalf("wrong value")
	}
	if m.GetBool("bool") != true {
		t.Fatalf("wrong value")
	}
	if m.GetBool("not exist key") != false {
		t.Fatalf("wrong value")
	}
	if m.GetStr("string") != "xxx" {
		t.Fatalf("wrong value")
	}
	if m.GetStr("not exist key") != "" {
		t.Fatalf("wrong value")
	}
	if math.Abs(m.GetFloat("float32")-32.0) >= 0.000001 {
		t.Fatalf("wrong value")
	}
	if math.Abs(m.GetFloat("float64")-64.0) >= 0.000001 {
		t.Fatalf("wrong value")
	}
	if math.Abs(m.GetFloat("not exist key")-0.0) >= 0.000001 {
		t.Fatalf("wrong value")
	}
	if !m.HasKey("bson.M") {
		t.Fatalf("should has key")
	}
	sm := m.GetMapAttr("bson.M")
	t.Logf("bson.M convert to %v", sm)
	sl := m.GetListAttr("list")
	t.Logf("list convert to %v", sl)
}
