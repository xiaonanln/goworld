package bigworld_netutil

import (
	"net"

	"runtime/debug"
	"time"

	"github.com/xiaonanln/vacuum/vlog"
)

const (
	RESTART_TCP_SERVER_INTERVAL = 3 * time.Second
)

type TCPServerDelegate interface {
	ServeTCPConnection(net.Conn)
}

func ServeTCPForever(listenAddr string, delegate TCPServerDelegate) {
	for {
		err := serveTCPForeverOnce(listenAddr, delegate)
		vlog.Error("server@%s failed with error: %v, will restart after %s", listenAddr, err, RESTART_TCP_SERVER_INTERVAL)
		time.Sleep(RESTART_TCP_SERVER_INTERVAL)
	}
}

func serveTCPForeverOnce(listenAddr string, delegate TCPServerDelegate) error {
	defer func() {
		if err := recover(); err != nil {
			vlog.Error("serveTCPImpl: paniced with error %s", err)
			debug.PrintStack()
		}
	}()

	return ServeTCP(listenAddr, delegate)

}

func ServeTCP(listenAddr string, delegate TCPServerDelegate) error {
	ln, err := net.Listen("tcp", listenAddr)
	vlog.Info("Listening on TCP: %s ...", listenAddr)

	if err != nil {
		return err
	}

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			} else {
				return err
			}
		}

		vlog.Info("Connection from: %s", conn.RemoteAddr())
		go delegate.ServeTCPConnection(conn)
	}
	return nil
}
