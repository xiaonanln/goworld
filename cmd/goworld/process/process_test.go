package process

import "testing"

func TestProcesses(t *testing.T) {
	ps, err := Processes()
	if err != nil {
		t.Errorf("ListProcess error: %s", err)
	}

	for _, p := range ps {
		cmdline, err := p.CmdlineSlice()
		t.Logf("process %s, err %v", cmdline, err)
	}
}
