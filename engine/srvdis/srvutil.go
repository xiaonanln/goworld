package srvdis

import (
	"fmt"
	"strings"
)

func registerPath(srvType, srvId string) string {
	return fmt.Sprintf("/srvdis/%s/%s", srvType, srvId)
}

func parseRegisterPath(key []byte) (srvtype, srvid string) {
	srvpath := string(key[len("/srvdis/"):])
	srvpathSp := strings.Split(srvpath, "/")
	srvtype = srvpathSp[0]
	srvid = srvpathSp[1]
	return
}
