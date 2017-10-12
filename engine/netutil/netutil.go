package netutil

import (
	"fmt"
	"io"
	"net"
	"reflect"

	"os"

	"unsafe"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// IsConnectionError check if the error is a connection error (close)
func IsConnectionError(_err interface{}) bool {
	err, ok := _err.(error)
	if !ok {
		return false
	}

	err = errors.Cause(err)
	if err == io.EOF {
		return true
	}

	neterr, ok := err.(net.Error)
	if !ok {
		return false
	}
	if neterr.Timeout() {
		return false
	}

	return true
}

// ConnectTCP connects to host:port in TCP
func ConnectTCP(host string, port int) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	return conn, err
}

// ServeForever runs the function with arguments forever
//
// ServeForever will restart the function call if function panics,
func ServeForever(f interface{}, args ...interface{}) {
	fval := reflect.ValueOf(f)
	argscount := len(args)
	argVals := make([]reflect.Value, argscount, argscount)
	for i := 0; i < argscount; i++ {
		argVals[i] = reflect.ValueOf(args[i])
	}

	for {
		runServe(fval, argVals)
		if consts.DEBUG_MODE { // we just quit in debug mode
			os.Exit(2)
		}
	}
}

func runServe(f reflect.Value, args []reflect.Value) {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.TraceError("ServeForever: func %v quited with error %v", f, err)
		}
	}()

	rets := f.Call(args)
	gwlog.Debugf("ServeForever: func %v returns %v", f, rets)
}

func PutFloat32(b []byte, f float32) {
	NETWORK_ENDIAN.PutUint32(b, *(*uint32)(unsafe.Pointer(&f)))
}
