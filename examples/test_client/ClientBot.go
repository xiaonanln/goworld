package main

import (
	"sync"

	"fmt"

	"math/rand"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/proto"
	"github.com/xiaonanln/vacuum/netutil"
)

type ClientBot struct {
	id     int
	waiter *sync.WaitGroup
	conn   proto.GoWorldConnection
}

func newClientBot(id int, waiter *sync.WaitGroup) *ClientBot {
	return &ClientBot{
		id:     id,
		waiter: waiter,
	}
}

func (bot *ClientBot) String() string {
	return fmt.Sprintf("ClientBot<%d>", bot.id)
}

func (bot *ClientBot) run() {
	defer bot.waiter.Done()

	gwlog.Info("%s is running ...", bot)

	serverIDs := config.GetServerIDs()
	// choose a random serverID
	serverID := serverIDs[rand.Intn(len(serverIDs))]
	serverID = 1 // lndebug
	gwlog.Debug("%s is connecting to server %d", bot, serverID)
	cfg := config.GetServer(serverID)
	cfg = cfg
	conn, err := netutil.ConnectTCP(cfg.Ip, cfg.Port)
	if err != nil {
		gwlog.Error("Connect failed: %s", err)
		return
	}
	gwlog.Info("connected: %s", conn.RemoteAddr())
	bot.conn = proto.NewGoWorldConnection(conn)
	defer bot.conn.Close()

	bot.loop()
}

func (bot *ClientBot) loop() {
	var msgtype proto.MsgType_t
	for {
		pkt, err := bot.conn.Recv(&msgtype)
		if err != nil {
			gwlog.Panic(err)
		}
		gwlog.Info("recv packet: %v", pkt.Payload())
	}
}
