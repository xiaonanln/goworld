package entity

import (
	"reflect"
	"strings"
)

const (
	RF_SERVER       = 1 << iota
	RF_OWN_CLIENT   = 1 << iota
	RF_OTHER_CLIENT = 1 << iota
)

type RpcDesc struct {
	Func       reflect.Value
	Flags      uint
	MethodType reflect.Type
	NumArgs    int
}

type RpcDescMap map[string]*RpcDesc

func (rdm RpcDescMap) visit(method reflect.Method) {
	methodName := method.Name
	var flag uint
	var rpcName string
	if strings.HasSuffix(methodName, "_Server") {
		// server method
		flag |= RF_SERVER
		rpcName = methodName[:len(methodName)-7]
	} else if strings.HasSuffix(methodName, "_Client") {
		flag |= (RF_SERVER + RF_OWN_CLIENT)
		rpcName = methodName[:len(methodName)-7]
	} else if strings.HasSuffix(methodName, "_AllClient") {
		flag |= (RF_SERVER + RF_OWN_CLIENT + RF_OTHER_CLIENT)
		rpcName = methodName[:len(methodName)-10]
	} else {
		// not a rpc method
		return
	}

	methodType := method.Type
	rdm[rpcName] = &RpcDesc{
		Func:       method.Func,
		Flags:      flag,
		MethodType: methodType,
		NumArgs:    methodType.NumIn() - 1, // do not count the receiver
	}
}
