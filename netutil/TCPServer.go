package netutil

import (
	"net"
	"time"

	"os"

	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
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
		gwlog.Error("server@%s failed with error: %v, will restart after %s", listenAddr, err, RESTART_TCP_SERVER_INTERVAL)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
		time.Sleep(RESTART_TCP_SERVER_INTERVAL)
	}
}

func serveTCPForeverOnce(listenAddr string, delegate TCPServerDelegate) error {
	defer func() {
		if err := recover(); err != nil {
			gwlog.TraceError("serveTCPImpl: paniced with error %s", err)
		}
	}()

	return ServeTCP(listenAddr, delegate)

}

func ServeTCP(listenAddr string, delegate TCPServerDelegate) error {
	ln, err := net.Listen("tcp", listenAddr)
	gwlog.Info("Listening on TCP: %s ...", listenAddr)

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

		gwlog.Info("Connection from: %s", conn.RemoteAddr())
		go delegate.ServeTCPConnection(conn)
	}
	return nil
}
