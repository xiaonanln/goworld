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
	n := len(srvpathSp)
	if n == 2 {
		srvtype = srvpathSp[0]
		srvid = srvpathSp[1]
	} else { // len(srvpathSp) > 2
		srvtype = strings.Join(srvpathSp[:n-1], "/")
		srvid = srvpathSp[n-1]
	}
	return
}
