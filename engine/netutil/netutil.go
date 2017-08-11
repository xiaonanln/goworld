package netutil

import (
	"fmt"
	"io"
	"net"
	"reflect"

	"os"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// IsTemporaryNetError checks if the error is a temporary network error
func IsTemporaryNetError(err error) bool {
	if err == nil {
		println("nil")
		return false
	}

	err = errors.Cause(err)
	netErr, ok := err.(net.Error)
	if !ok {
		return false
	}
	return netErr.Temporary() || netErr.Timeout()
}

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
	if neterr.Temporary() || neterr.Timeout() {
		return false
	}

	return true
}

// WriteAll write all bytes of data to the writer
func WriteAll(conn io.Writer, data []byte) error {
	left := len(data)
//	olen := left
//	wc := 0
//	defer func(){
//		fmt.Fprintf(os.Stderr, "(WA%d/%d)", olen, wc)
//	}()
	for left > 0 {
		n, err := conn.Write(data)
		//wc += 1
		if n == left && err == nil { // handle most common case first
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil && !IsTemporaryNetError(err) {
			return err
		}
	}
	return nil
}

// ReadAll reads from the reader until all bytes in data is filled
func ReadAll(conn io.Reader, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := conn.Read(data)
		if n == left && err == nil { // handle most common case first
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil && !IsTemporaryNetError(err) {
			return err
		}
	}
	return nil
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
