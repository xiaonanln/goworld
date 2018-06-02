package netutil

import (
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwioutil"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	_RESTART_TCP_SERVER_INTERVAL = 3 * time.Second
	_RESTART_UDP_SERVER_INTERVAL = 3 * time.Second
)

// TCPServerDelegate is the implementations that a TCP server should provide
type TCPServerDelegate interface {
	ServeTCPConnection(net.Conn)
}

// ServeTCPForever serves on specified address as TCP server, for ever ...
func ServeTCPForever(listenAddr string, delegate TCPServerDelegate) {
	for {
		serveTCPForeverOnce(listenAddr, delegate)
		//gwlog.Errorf("server@%s failed with error: %v, will restart after %s", listenAddr, err, _RESTART_TCP_SERVER_INTERVAL)
		time.Sleep(_RESTART_TCP_SERVER_INTERVAL)
	}
}

func serveTCPForeverOnce(listenAddr string, delegate TCPServerDelegate) {
	defer func() {
		if err := recover(); err != nil {
			gwlog.TraceError("ServeTCPForever: panic with error %s", err)
		}
	}()

	ServeTCP(listenAddr, delegate)
}

// ServeTCP serves on specified address as TCP server
func ServeTCP(listenAddr string, delegate TCPServerDelegate) {
	ln, err := net.Listen("tcp", listenAddr)
	gwlog.Infof("Listening on TCP: %s ...", listenAddr)

	if err != nil {
		gwlog.Fatal(errors.Wrap(err, "tcp listen failed"))
	}

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if gwioutil.IsTimeoutError(err) {
				continue
			} else {
				gwlog.Panic(errors.Wrap(err, "tcp accept failed"))
			}
		}

		gwlog.Infof("Connection from: %s", conn.RemoteAddr())
		go delegate.ServeTCPConnection(conn)
	}
}
