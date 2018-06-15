package process

import "testing"

func TestProcesses(t *testing.T) {
	ps, err := Processes()
	if err != nil {
		t.Errorf("ListProcess error: %s", err)
	}

	for _, p := range ps {
		exe, err := p.Path()
		t.Logf("process %s, err %v", exe, err)
	}
}
