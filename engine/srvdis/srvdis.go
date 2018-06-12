package srvdis

type RegisterCallback func(ok bool)

func Register(srvtype string, srvid string, info string, cb RegisterCallback) {

}
