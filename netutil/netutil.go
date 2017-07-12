package netutil

import (
	"fmt"
	"io"
	"net"
	"reflect"

	"os"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
)

func init() {
}

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

func IsConnectionClosed(_err interface{}) bool {
	err, ok := _err.(error)
	if !ok {
		return false
	}

	err = errors.Cause(err)
	if err == io.EOF {
		return true
	}

	neterr, ok := _err.(net.Error)
	if !ok {
		return false
	}
	if neterr.Temporary() || neterr.Timeout() {
		return false
	}

	return true
}

func WriteAll(conn io.Writer, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := conn.Write(data)
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

func ReadLine(conn net.Conn) (string, error) {
	var _linebuff [1024]byte
	linebuff := _linebuff[0:0]

	buff := [1]byte{0} // buff of just 1 byte

	for {
		n, err := conn.Read(buff[0:1])
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			} else {
				return "", err
			}
		}
		if n == 1 {
			c := buff[0]
			if c == '\n' {
				return string(linebuff), nil
			} else {
				linebuff = append(linebuff, c)
			}
		}
	}
}

func ConnectTCP(host string, port int) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	return conn, err
}

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
	gwlog.Debug("ServeForever: func %v returns %v", f, rets)
}
