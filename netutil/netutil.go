package netutil

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"runtime/debug"

	"github.com/xiaonanln/goworld/gwlog"
)

func init() {
}

func IsTemporaryNetError(err error) bool {
	if err == nil {
		return false
	}

	netErr, ok := err.(net.Error)
	if !ok {
		return false
	}
	return netErr.Temporary() || netErr.Timeout()
}

func IsConnectionClosed(_err interface{}) bool {
	err, ok := _err.(error)
	if ok && err == io.EOF {
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

func WriteAll(conn net.Conn, data []byte) error {
	for len(data) > 0 {
		n, err := conn.Write(data)
		if n > 0 {
			data = data[n:]
		}
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			} else {
				return err
			}
		}
	}
	return nil
}

func ReadAll(conn net.Conn, data []byte) error {
	for len(data) > 0 {
		n, err := conn.Read(data)
		if n > 0 {
			data = data[n:]
		}
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			} else {
				return err
			}
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
	}
}

func runServe(f reflect.Value, args []reflect.Value) {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.Error("ServeForever: func %v quited with error %v", f, err)
			debug.PrintStack()
		}
	}()

	rets := f.Call(args)
	gwlog.Debug("ServeForever: func %v returns %v", f, rets)
}

func RecvAll(conn io.Reader, buf []byte) error {
	for len(buf) > 0 {
		n, err := conn.Read(buf)
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			}

			return err
		}
		buf = buf[n:]
	}
	return nil
}

func SendAll(conn io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			}

			return err
		}
		data = data[n:]
	}
	return nil
}
