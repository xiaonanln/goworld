package dispatchercluster

import (
	"hash/fnv"

	"github.com/xiaonanln/goworld/engine/common"
)

func hashEntityID(id common.EntityID) int {
	// hash EntityID to dispatcher shard index: use least 2 bytes
	b1 := id[14]
	b2 := id[15]
	return int(b1)*256 + int(b2)
}

func hashGateID(gateid uint16) int {
	return int(gateid - 1)
}

func hashServiceName(name string) int {
	h := fnv.New32a()
	h.Write([]byte(name))
	return int(h.Sum32())
}
