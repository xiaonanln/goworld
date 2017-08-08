package common

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestStringSet(t *testing.T) {
	ss := StringSet{}
	ss.Add("1")
	ss.Add("2")
	assert.T(t, ss.Contains("1"), "should contain")
	assert.T(t, ss.Contains("2"), "should contain")
	ss.Remove("2")
	assert.T(t, !ss.Contains("2"), "should contain")
}

func TestStringList(t *testing.T) {
	ss := StringList{}
	ss.Append("1")
	assert.T(t, len(ss) == 1, "wrong length")
	ss.Append("2")
	assert.T(t, len(ss) == 2, "wrong length")
	ss.Append("3")
	assert.T(t, len(ss) == 3, "wrong length")
	ss.Remove("2")
	assert.Tf(t, len(ss) == 2, "wrong length: %v", ss)
	assert.Tf(t, ss.Find("1") == 0, "wrong index: %d", ss.Find("1"))
	assert.Tf(t, ss.Find("2") == -1, "wrong index: %d", ss.Find("2"))
	assert.Tf(t, ss.Find("3") == 1, "wrong index: %d", ss.Find("3"))
}
