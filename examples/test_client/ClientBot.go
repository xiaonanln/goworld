package main

import (
	"sync"

	"fmt"

	"math/rand"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type ClientBot struct {
	id       int
	waiter   *sync.WaitGroup
	conn     proto.GoWorldConnection
	entities map[common.EntityID]*ClientEntity

	logined bool
}

func newClientBot(id int, waiter *sync.WaitGroup) *ClientBot {
	return &ClientBot{
		id:       id,
		waiter:   waiter,
		entities: map[common.EntityID]*ClientEntity{},
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
		//gwlog.Info("recv packet: msgtype=%v, packet=%v", msgtype, pkt.Payload())
		bot.handlePacket(msgtype, pkt)
	}
}
func (bot *ClientBot) handlePacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	_ = packet.ReadUint16()
	_ = packet.ReadClientID()
	if msgtype == proto.MT_CREATE_ENTITY_ON_CLIENT {
		typeName := packet.ReadVarStr()
		entityid := packet.ReadEntityID()
		gwlog.Info("Create entity %s.%s", typeName, entityid)
		bot.createEntity(typeName, entityid)
	} else if msgtype == proto.MT_DESTROY_ENTITY_ON_CLIENT {
		typeName := packet.ReadVarStr()
		entityid := packet.ReadEntityID()
		gwlog.Info("Destroy entity %s.%s", typeName, entityid)
		bot.destroyEntity(typeName, entityid)
	}
}

func (bot *ClientBot) createEntity(typeName string, entityid common.EntityID) {
	if bot.entities[entityid] == nil {
		e := newClientEntity(bot, typeName, entityid)
		bot.entities[entityid] = e
	}
}

func (bot *ClientBot) destroyEntity(typeName string, entityid common.EntityID) {
	if bot.entities[entityid] != nil {
		delete(bot.entities, entityid)
	}
}

func (bot *ClientBot) username() string {
	return fmt.Sprintf("test%d", bot.id)
}

func (bot *ClientBot) password() string {
	return "123456"
}

func (bot *ClientBot) CallServer(id common.EntityID, method string, args []interface{}) {
	gwlog.Info("%s call server: %s.%s%v", bot, id, method, args)
	bot.conn.SendCallEntityMethodFromClient(id, method, args)
}
