package main

import (
	"net"
	"sync"

	"fmt"

	"math/rand"

	"time"

	"reflect"

	"os"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/post"
	"github.com/xiaonanln/goworld/proto"
)

type ClientBot struct {
	sync.Mutex

	id                 int
	waiter             *sync.WaitGroup
	conn               *proto.GoWorldConnection
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

	gateIDs := config.GetGateIDs()
	// choose a random gateid
	gateid := gateIDs[rand.Intn(len(gateIDs))]
	gwlog.Debug("%s is connecting to gate %d", bot, gateid)
	cfg := config.GetGate(gateid)
	cfg = cfg
	var netconn net.Conn
	var err error
	for { // retry for ever
		//netconn, err = netutil.ConnectTCP("10.246.13.148", cfg.Port)
		netconn, err = netutil.ConnectTCP("localhost", cfg.Port)
		if err != nil {
			gwlog.Error("Connect failed: %s", err)
			time.Sleep(time.Second * time.Duration(1+rand.Intn(10)))
			continue
		}
		// connected , ok
		break
	}
	netconn.(*net.TCPConn).SetWriteBuffer(64 * 1024)
	netconn.(*net.TCPConn).SetReadBuffer(64 * 1024)
	gwlog.Info("connected: %s", netconn.RemoteAddr())

	var conn netutil.Connection = netutil.NetConnection{netconn}
	conn = netutil.NewBufferedReadConnection(conn)
	//if cfg.CompressConnection {
	//	conn = netutil.NewCompressedConnection(conn)
	//}

	bot.conn = proto.NewGoWorldConnection(conn, cfg.CompressConnection)
	defer bot.conn.Close()

	bot.loop()
}

func (bot *ClientBot) loop() {
	var msgtype proto.MsgType_t
	for {
		err := bot.conn.SetRecvDeadline(time.Now().Add(time.Millisecond * 100))
		if err != nil {
			gwlog.Panic(err)
		}

		pkt, err := bot.conn.Recv(&msgtype)

		if pkt != nil {
			bot.handlePacket(msgtype, pkt)
			pkt.Release()
		} else if err != nil && !netutil.IsTemporaryNetError(err) {
			// bad error
			gwlog.Panic(err)
		}

		bot.conn.Flush()
		post.Tick()
	}
}

func (bot *ClientBot) handlePacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	bot.Lock()
	defer bot.Unlock()

	if msgtype != proto.MT_CALL_FILTERED_CLIENTS {
		_ = packet.ReadUint16()
		_ = packet.ReadClientID() // TODO: strip these two fields ? seems a little difficult, maybe later.
	}

	if msgtype == proto.MT_NOTIFY_ATTR_CHANGE_ON_CLIENT {
		entityID := packet.ReadEntityID()
		path := packet.ReadStringList()
		key := packet.ReadVarStr()
		var val interface{}
		packet.ReadData(&val)
		if !quiet {
			gwlog.Debug("Entity %s Attribute %v: set %s=%v", entityID, path, key, val)
		}
		bot.applyAttrChange(entityID, path, key, val)
	} else if msgtype == proto.MT_NOTIFY_ATTR_DEL_ON_CLIENT {
		entityID := packet.ReadEntityID()
		path := packet.ReadStringList()
		key := packet.ReadVarStr()
		if !quiet {
			gwlog.Debug("Entity %s Attribute %v deleted %s", entityID, path, key)
		}
		bot.applyAttrDel(entityID, path, key)
	} else if msgtype == proto.MT_CREATE_ENTITY_ON_CLIENT {
		isPlayer := packet.ReadBool()
		entityID := packet.ReadEntityID()
		typeName := packet.ReadVarStr()

		x := entity.Coord(packet.ReadFloat32())
		y := entity.Coord(packet.ReadFloat32())
		z := entity.Coord(packet.ReadFloat32())
		yaw := entity.Yaw(packet.ReadFloat32())
		//gwlog.Info("Create entity %s.%s: isPlayer=%v", typeName, entityID, isPlayer)
		var clientData map[string]interface{}
		packet.ReadData(&clientData)

		if typeName == entity.SPACE_ENTITY_TYPE {
			// this is a space
			bot.createSpace(entityID, clientData)
		} else {
			// this is a entity
			bot.createEntity(typeName, entityID, isPlayer, clientData, x, y, z, yaw)
		}
	} else if msgtype == proto.MT_DESTROY_ENTITY_ON_CLIENT {
		typeName := packet.ReadVarStr()
		entityID := packet.ReadEntityID()
		if !quiet {
			gwlog.Debug("Destroy entity %s.%s", typeName, entityID)
		}
		if typeName == entity.SPACE_ENTITY_TYPE {
			bot.destroySpace(entityID)
		} else {
			bot.destroyEntity(typeName, entityID)
		}
	} else if msgtype == proto.MT_CALL_ENTITY_METHOD_ON_CLIENT {
		entityID := packet.ReadEntityID()
		method := packet.ReadVarStr()
		args := packet.ReadArgs()
		if !quiet {
			gwlog.Debug("Call entity %s.%s(%v)", entityID, method, args)
		}
		bot.callEntityMethod(entityID, method, args)
	} else if msgtype == proto.MT_CALL_FILTERED_CLIENTS {
		_ = packet.ReadVarStr() // ignore key
		_ = packet.ReadVarStr() // ignore val
		method := packet.ReadVarStr()
		args := packet.ReadArgs()
		if bot.player == nil {
			gwlog.Warn("Player not found while calling filtered client")
			return
		}

		bot.callEntityMethod(bot.player.ID, method, args)
	} else if msgtype == proto.MT_UPDATE_POSITION_ON_CLIENT {
		entityID := packet.ReadEntityID()
		x := entity.Coord(packet.ReadFloat32())
		y := entity.Coord(packet.ReadFloat32())
		z := entity.Coord(packet.ReadFloat32())
		bot.updateEntityPosition(entityID, entity.Position{x, y, z})
	} else if msgtype == proto.MT_UPDATE_YAW_ON_CLIENT {
		entityID := packet.ReadEntityID()
		yaw := entity.Yaw(packet.ReadFloat32())
		bot.updateEntityYaw(entityID, yaw)
	} else {
		gwlog.Panicf("unknown msgtype: %v", msgtype)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
	}
}

func (bot *ClientBot) updateEntityPosition(entityID common.EntityID, position entity.Position) {
	//gwlog.Debug("updateEntityPosition %s => %s", entityID, position)
	if bot.entities[entityID] == nil {
		gwlog.Error("entity %s not found")
		return
	}
	entity := bot.entities[entityID]
	entity.pos = position
	entity.onUpdatePosition()
}

func (bot *ClientBot) updateEntityYaw(entityID common.EntityID, yaw entity.Yaw) {
	//gwlog.Debug("updateEntityYaw %s => %s", entityID, yaw)
	if bot.entities[entityID] == nil {
		gwlog.Error("entity %s not found")
		return
	}
	entity := bot.entities[entityID]
	entity.yaw = yaw
	entity.onUpdateYaw()
}

func (bot *ClientBot) applyAttrChange(entityID common.EntityID, path []string, key string, val interface{}) {
	//gwlog.Info("SET ATTR %s.%v: set %s=%v", entityID, path, key, val)
	if bot.entities[entityID] == nil {
		gwlog.Error("entity %s not found")
		return
	}
	entity := bot.entities[entityID]
	entity.applyAttrChange(path, key, val)
}

func (bot *ClientBot) applyAttrDel(entityID common.EntityID, path []string, key string) {
	//gwlog.Info("DEL ATTR %s.%v: del %s", entityID, path, key)
	if bot.entities[entityID] == nil {
		gwlog.Error("entity %s not found")
		return
	}
	entity := bot.entities[entityID]
	entity.applyAttrDel(path, key)
}

func (bot *ClientBot) createEntity(typeName string, entityID common.EntityID, isPlayer bool, clientData map[string]interface{}, x, y, z entity.Coord, yaw entity.Yaw) {
	if bot.entities[entityID] == nil {
		e := newClientEntity(bot, typeName, entityID, isPlayer, clientData, x, y, z, yaw)
		bot.entities[entityID] = e
		if isPlayer {
			if bot.player != nil {
				gwlog.TraceError("%s.createEntity: creating player %S, but player is already set to %s", bot, e, bot.player)
			}
			bot.player = e
		}
	}
}

func (bot *ClientBot) destroyEntity(typeName string, entityID common.EntityID) {
	entity := bot.entities[entityID]
	if entity != nil {
		entity.Destroy()
		if entity == bot.player {
			bot.player = nil
		}
		delete(bot.entities, entityID)
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

func (bot *ClientBot) callEntityMethod(entityID common.EntityID, method string, args [][]byte) {
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
		argValPtr := reflect.New(argType)
		netutil.MSG_PACKER.UnpackMsg(arg, argValPtr.Interface())
		in[i] = reflect.Indirect(argValPtr)
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
