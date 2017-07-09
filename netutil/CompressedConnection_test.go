package netutil

import (
	"net"

	"testing"

	"sync"

	"github.com/xiaonanln/goworld/gwlog"
)

var (
	listener net.Listener
)

func init() {
	var err error
	listener, err = net.Listen("tcp", "0.0.0.0:13001")
	paniciferror(err)
	gwlog.Info("Listening ...")
}

func paniciferror(err error) {
	if err != nil {
		gwlog.Panicf("error: %s", err.Error())
	}
}

func TestCompressedConnection_Read(t *testing.T) {
	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		conn, err := net.Dial("tcp", "localhost:13001")
		println("dial", conn, err)
		paniciferror(err)
		cc := NewCompressedConnection(NetConnection{conn})
		println("NewCompressedConnection", cc)
		n, err := cc.Write([]byte("hello"))
		println("write", n, err)
		err = cc.Flush()
		paniciferror(err)
		println("flush", err)
		cc.Write([]byte("world"))
		cc.Flush()
		cc.Write([]byte("bigaa"))
		cc.Flush()

		cc.Write([]byte("bigha"))
		cc.Flush()
	}()
	go func() {
		defer wait.Done()

		conn, err := listener.Accept()
		println("accept", conn, err)
		paniciferror(err)
		cc := NewCompressedConnection(NetConnection{conn})
		println("NewCompressedConnection", cc)
		buf := make([]byte, 5)
		n, err := cc.Read(buf)
		paniciferror(err)
		println("read", n, err, string(buf[:n]))
		if string(buf[:n]) != "hello" {
			gwlog.Panicf("recv error")
		}

		n, err = cc.Read(buf)
		paniciferror(err)
		println("read", n, err, string(buf[:n]))
		if string(buf[:n]) != "world" {
			gwlog.Panicf("recv error")
		}
		buf = make([]byte, 3)
		n, err = cc.Read(buf)
		paniciferror(err)
		println("read", n, err, string(buf[:n]))
		if string(buf[:n]) != "big" {
			gwlog.Panicf("recv error")
		}
		n, err = cc.Read(buf)
		paniciferror(err)
		println("read", n, err, string(buf[:n]))
		if string(buf[:n]) != "aa" {
			gwlog.Panicf("recv error")
		}
		n, err = cc.Read(buf)
		paniciferror(err)
		println("read", n, err, string(buf[:n]))
		if string(buf[:n]) != "big" {
			gwlog.Panicf("recv error")
		}
		n, err = cc.Read(buf)
		paniciferror(err)
		println("read", n, err, string(buf[:n]))
		if string(buf[:n]) != "ha" {
			gwlog.Panicf("recv error")
		}
	}()

	wait.Wait()
}
