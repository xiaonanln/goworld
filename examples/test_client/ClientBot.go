package main

import (
	"net"
	"sync"

	"github.com/xiaonanln/typeconv"

	"fmt"

	"math/rand"

	"time"

	"reflect"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
	"github.com/xiaonanln/goworld/consts"
	"os"
)

type ClientBot struct {
	sync.Mutex

	id                 int
	waiter             *sync.WaitGroup
	conn               proto.GoWorldConnection
	entities           map[common.EntityID]*ClientEntity
	player             *ClientEntity
	currentSpace       *ClientSpace
	logined            bool
	startedDoingThings bool
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
	gwlog.Debug("%s is connecting to server %d", bot, serverID)
	cfg := config.GetServer(serverID)
	cfg = cfg
	var conn net.Conn
	var err error
	for { // retry for ever
		conn, err = netutil.ConnectTCP("localhost", cfg.Port)
		if err != nil {
			gwlog.Error("Connect failed: %s", err)
			time.Sleep(time.Second * time.Duration(1+rand.Intn(10)))
			continue
		}
		// connected , ok
		break
	}
	conn.(*net.TCPConn).SetWriteBuffer(64 * 1024)
	conn.(*net.TCPConn).SetReadBuffer(64 * 1024)
	gwlog.Info("connected: %s", conn.RemoteAddr())
	bot.conn = proto.NewGoWorldConnection(conn, false)
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
		pkt.Release()
	}
}

func (bot *ClientBot) handlePacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	bot.Lock()
	defer bot.Unlock()

	_ = packet.ReadUint16()
	_ = packet.ReadClientID() // TODO: strip these two fields ?
	if msgtype == proto.MT_NOTIFY_ATTR_CHANGE_ON_CLIENT {
		entityid := packet.ReadEntityID()
		path := packet.ReadStringList()
		key := packet.ReadVarStr()
		var val interface{}
		packet.ReadData(&val)
		if !quiet {
			gwlog.Debug("Entity %s Attribute %v: set %s=%v", entityid, path, key, val)
		}
		bot.applyAttrChange(entityid, path, key, val)
	} else if msgtype == proto.MT_NOTIFY_ATTR_DEL_ON_CLIENT {
		entityid := packet.ReadEntityID()
		path := packet.ReadStringList()
		key := packet.ReadVarStr()
		if !quiet {
			gwlog.Debug("Entity %s Attribute %v deleted %s", entityid, path, key)
		}
		bot.applyAttrDel(entityid, path, key)
	} else if msgtype == proto.MT_CREATE_ENTITY_ON_CLIENT {
		isPlayer := packet.ReadBool()
		entityid := packet.ReadEntityID()
		typeName := packet.ReadVarStr()

		var clientData map[string]interface{}
		packet.ReadData(&clientData)
		if !quiet {
			gwlog.Debug("Create entity %s.%s: isPlayer=%v, attrs=%v", typeName, entityid, isPlayer, clientData)
		}

		if typeName == entity.SPACE_ENTITY_TYPE {
			// this is a space
			bot.createSpace(entityid, clientData)
		} else {
			// this is a entity
			bot.createEntity(typeName, entityid, isPlayer, clientData)
		}
	} else if msgtype == proto.MT_DESTROY_ENTITY_ON_CLIENT {
		typeName := packet.ReadVarStr()
		entityid := packet.ReadEntityID()
		if !quiet {
			gwlog.Debug("Destroy entity %s.%s", typeName, entityid)
		}
		if typeName == entity.SPACE_ENTITY_TYPE {
			bot.destroySpace(entityid)
		} else {
			bot.destroyEntity(typeName, entityid)
		}
	} else if msgtype == proto.MT_CALL_ENTITY_METHOD_ON_CLIENT {
		entityID := packet.ReadEntityID()
		method := packet.ReadVarStr()
		var args []interface{}
		packet.ReadData(&args)
		if !quiet {
			gwlog.Debug("Call entity %s.%s(%v)", entityID, method, args)
		}
		bot.callEntityMethod(entityID, method, args)
	} else {
		gwlog.Panicf("unknown msgtype: %v", msgtype)
		if consts.DEBUG_MODE{
			os.Exit(2)
		}
	}
}

func (bot *ClientBot) applyAttrChange(entityid common.EntityID, path []string, key string, val interface{}) {
	//gwlog.Info("SET ATTR %s.%v: set %s=%v", entityid, path, key, val)
	if bot.entities[entityid] == nil {
		gwlog.Error("entity %s not found")
	}
	entity := bot.entities[entityid]
	entity.applyAttrChange(path, key, val)
}

func (bot *ClientBot) applyAttrDel(entityid common.EntityID, path []string, key string) {
	//gwlog.Info("DEL ATTR %s.%v: del %s", entityid, path, key)
	if bot.entities[entityid] == nil {
		gwlog.Error("entity %s not found")
	}
	entity := bot.entities[entityid]
	entity.applyAttrDel(path, key)
}

func (bot *ClientBot) createEntity(typeName string, entityid common.EntityID, isPlayer bool, clientData map[string]interface{}) {
	if bot.entities[entityid] == nil {
		e := newClientEntity(bot, typeName, entityid, isPlayer, clientData)
		bot.entities[entityid] = e
		if isPlayer {
			if bot.player != nil {
				gwlog.TraceError("%s.createEntity: creating player %S, but player is already set to %s", bot, e, bot.player)
			}
			bot.player = e
		}
	}
}

func (bot *ClientBot) destroyEntity(typeName string, entityid common.EntityID) {
	entity := bot.entities[entityid]
	if entity != nil {
		entity.Destroy()
		if entity == bot.player {
			bot.player = nil
		}
		delete(bot.entities, entityid)
	}
}

func (bot *ClientBot) createSpace(spaceID common.EntityID, data map[string]interface{}) {
	if bot.currentSpace != nil {
		gwlog.TraceError("%s.createSpace: duplicate space: %s and %s", bot, bot.currentSpace, spaceID)
	}
	space := newClientSpace(bot, spaceID, data)
	bot.currentSpace = space
	gwlog.Debug("%s current space change to %s", bot, space)
	bot.OnEnterSpace()
}

func (bot *ClientBot) destroySpace(spaceID common.EntityID) {
	if bot.currentSpace == nil || bot.currentSpace.ID != spaceID {
		gwlog.TraceError("%s.destroySpace: space %s not exists, current space is %s", bot, spaceID, bot.currentSpace)
		return
	}
	oldSpace := bot.currentSpace
	bot.currentSpace = nil
	gwlog.Debug("%s: leave current space %s", bot, spaceID)
	bot.OnLeaveSpace(oldSpace)
}

func (bot *ClientBot) callEntityMethod(entityID common.EntityID, method string, args []interface{}) {
	entity := bot.entities[entityID]
	if entity == nil {
		gwlog.Warn("Entity %s is not found while calling method %s(%v)", entityID, method, args)
		return
	}

	methodVal := reflect.ValueOf(entity).MethodByName(method)
	if !methodVal.IsValid() {
		gwlog.Error("Client method %s is not found", method)
		return
	}

	methodType := methodVal.Type()
	in := make([]reflect.Value, len(args))

	for i, arg := range args {
		argType := methodType.In(i)
		in[i] = typeconv.Convert(arg, argType)
	}
	methodVal.Call(in)
}

func (bot *ClientBot) username() string {
	return fmt.Sprintf("test%d", bot.id)
}

func (bot *ClientBot) password() string {
	return "123456"
}

func (bot *ClientBot) CallServer(id common.EntityID, method string, args []interface{}) {
	if !quiet {
		gwlog.Debug("%s call server: %s.%s%v", bot, id, method, args)
	}
	bot.conn.SendCallEntityMethodFromClient(id, method, args)
}

func (bot *ClientBot) OnEnterSpace() {
	gwlog.Debug("%s.OnEnterSpace, player=%s", bot, bot.player)
	player := bot.player
	if !bot.startedDoingThings {
		bot.startedDoingThings = true
		player.doSomethingLater()
	} else {
		player.notifyThingDone("DoEnterRandomSpace")
	}
}

func (bot *ClientBot) OnLeaveSpace(oldSpace *ClientSpace) {
	gwlog.Debug("%s.OnLeaveSpace, player=%s", bot, bot.player)
}
