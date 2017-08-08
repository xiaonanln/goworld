package common

import "testing"

func TestEntityID(t *testing.T) {
	eid := GenEntityID()
	if len(eid) != ENTITYID_LENGTH {
		t.Fail()
	}

	if eid.IsNil() {
		t.Fail()
	}

	if !EntityID("").IsNil() {
		t.Fail()
	}

}

func TestClientID(t *testing.T) {
	if !ClientID("").IsNil() {
		t.Fail()
	}
	cid := GenClientID()
	if cid.IsNil() {
		t.Fail()
	}
	if len(cid) != CLIENTID_LENGTH {
		t.Fail()
	}

}
