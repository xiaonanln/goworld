package main

import (
	"net"

	"os"
	"time"

	"github.com/xiaonanln/goworld/netutil"
)

func main() {
	//_ = make(chan int, 1024*1024*1024*3)
	conn, err := netutil.ConnectTCP("127.0.0.1", 13000)
	if err != nil {
		panic(err)
	}
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetWriteBuffer(1024 * 1024 * 1024 * 4)
	tcpConn.Write([]byte{'a'})
	for {
		os.Stdout.WriteString(".")
		time.Sleep(time.Second)
	}
}
