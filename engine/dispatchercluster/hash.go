package dispatchercluster

import (
	"github.com/syndtr/goleveldb/leveldb/util"
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

func hashString(s string) int {
	h := util.Hash([]byte(s), 0xbc9f1d34)
	return int(h)
}

func hashSrvID(sn string) int {
	return hashString(sn)
}
