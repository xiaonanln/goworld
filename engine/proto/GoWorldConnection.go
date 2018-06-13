package proto

import (
	"net"

	"time"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/netutil/compress"
)

// GoWorldConnection is the network protocol implementation of GoWorld components (dispatcher, gate, game)
type GoWorldConnection struct {
	packetConn   *netutil.PacketConnection
	closed       xnsyncutil.AtomicBool
	autoFlushing bool
}

// NewGoWorldConnection creates a GoWorldConnection using network connection
func NewGoWorldConnection(conn netutil.Connection, compressConnection bool, compressFormat string) *GoWorldConnection {
	var compressor compress.Compressor
	if compressConnection {
		compressor = compress.NewCompressor(compressFormat)
	}

	return &GoWorldConnection{
		packetConn: netutil.NewPacketConnection(conn, compressor),
	}
}

// SendSetGameID sends MT_SET_GAME_ID message
func (gwc *GoWorldConnection) SendSetGameID(id uint16, isReconnect bool, isRestore bool, isBanBootEntity bool,
	eids []common.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GAME_ID)
	packet.AppendUint16(id)
	packet.AppendBool(isReconnect)
	packet.AppendBool(isRestore)
	packet.AppendBool(isBanBootEntity)
	// put all entity IDs to the packet
	// TODO: use AppendEntityIDList
	packet.AppendUint32(uint32(len(eids)))
	for _, eid := range eids {
		packet.AppendEntityID(eid)
	}
	return gwc.SendPacketRelease(packet)
}

// SendSetGateID sends MT_SET_GATE_ID message
func (gwc *GoWorldConnection) SendSetGateID(id uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GATE_ID)
	packet.AppendUint16(id)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyCreateEntity sends MT_NOTIFY_CREATE_ENTITY message
func (gwc *GoWorldConnection) SendNotifyCreateEntity(id common.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CREATE_ENTITY)
	packet.AppendEntityID(id)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyDestroyEntity sends MT_NOTIFY_DESTROY_ENTITY message
func (gwc *GoWorldConnection) SendNotifyDestroyEntity(id common.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_DESTROY_ENTITY)
	packet.AppendEntityID(id)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyClientConnected sends MT_NOTIFY_CLIENT_CONNECTED message
func (gwc *GoWorldConnection) SendNotifyClientConnected(id common.ClientID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_CONNECTED)
	packet.AppendClientID(id)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyClientDisconnected sends MT_NOTIFY_CLIENT_DISCONNECTED message
func (gwc *GoWorldConnection) SendNotifyClientDisconnected(id common.ClientID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_DISCONNECTED)
	packet.AppendClientID(id)
	return gwc.SendPacketRelease(packet)
}

// SendCreateEntityAnywhere sends MT_CREATE_ENTITY_ANYWHERE message
func (gwc *GoWorldConnection) SendCreateEntityAnywhere(entityid common.EntityID, typeName string, data map[string]interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ANYWHERE)
	packet.AppendEntityID(entityid)
	packet.AppendVarStr(typeName)
	packet.AppendData(data)
	return gwc.SendPacketRelease(packet)
}

// SendLoadEntityAnywhere sends MT_LOAD_ENTITY_ANYWHERE message
func (gwc *GoWorldConnection) SendLoadEntityAnywhere(typeName string, entityID common.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_LOAD_ENTITY_ANYWHERE)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(typeName)
	return gwc.SendPacketRelease(packet)
}

// SendSrvdisRegister
func (gwc *GoWorldConnection) SendSrvdisRegister(srvid string, info string, force bool) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SRVDIS_REGISTER)
	packet.AppendVarStr(srvid)
	packet.AppendVarStr(info)
	packet.AppendBool(force)
	return gwc.SendPacketRelease(packet)
}

// SendCallEntityMethod sends MT_CALL_ENTITY_METHOD message
func (gwc *GoWorldConnection) SendCallEntityMethod(id common.EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	return gwc.SendPacketRelease(packet)
}

// SendCallEntityMethodFromClient sends MT_CALL_ENTITY_METHOD_FROM_CLIENT message
func (gwc *GoWorldConnection) SendCallEntityMethodFromClient(id common.EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_FROM_CLIENT)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	return gwc.SendPacketRelease(packet)
}

// SendCreateEntityOnClient sends MT_CREATE_ENTITY_ON_CLIENT message
func (gwc *GoWorldConnection) SendCreateEntityOnClient(gameid uint16, clientid common.ClientID, typeName string, entityid common.EntityID,
	isPlayer bool, clientData map[string]interface{}, x, y, z float32, yaw float32) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ON_CLIENT)
	packet.AppendUint16(gameid)
	packet.AppendClientID(clientid)
	packet.AppendBool(isPlayer)
	packet.AppendEntityID(entityid)
	packet.AppendVarStr(typeName)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	packet.AppendFloat32(yaw)
	packet.AppendData(clientData)
	return gwc.SendPacketRelease(packet)
}

// SendSyncPositionYawFromClient sends MT_SYNC_POSITION_YAW_FROM_CLIENT message
func (gwc *GoWorldConnection) SendSyncPositionYawFromClient(entityID common.EntityID, x, y, z float32, yaw float32) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SYNC_POSITION_YAW_FROM_CLIENT)
	packet.AppendEntityID(entityID)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	packet.AppendFloat32(yaw)
	return gwc.SendPacketRelease(packet)
}

//func (gwc *GoWorldConnection) SendSetClientClientID(clientid common.ClientID) error {
//	packet := gwc.packetConn.NewPacket()
//	packet.AppendUint16(MT_SET_CLIENT_CLIENTID)
//	packet.AppendClientID(clientid)
//	return gwc.SendPacketRelease(packet)
//}

func (gwc *GoWorldConnection) SetHeartbeatFromClient() error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_HEARTBEAT_FROM_CLIENT)
	return gwc.SendPacketRelease(packet)

}

// SendDestroyEntityOnClient sends MT_DESTROY_ENTITY_ON_CLIENT message
func (gwc *GoWorldConnection) SendDestroyEntityOnClient(gateid uint16, clientid common.ClientID, typeName string, entityid common.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DESTROY_ENTITY_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(typeName)
	packet.AppendEntityID(entityid)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyMapAttrChangeOnClient sends MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrChangeOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, key string, val interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendVarStr(key)
	packet.AppendData(val)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyMapAttrDelOnClient sends MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrDelOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, key string) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendVarStr(key)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyMapAttrClearOnClient sends MT_NOTIFY_MAP_ATTR_CLEAR_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrClearOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_CLEAR_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyListAttrChangeOnClient sends MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrChangeOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, index uint32, val interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendUint32(index)
	packet.AppendData(val)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyListAttrPopOnClient sends MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrPopOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	return gwc.SendPacketRelease(packet)
}

// SendNotifyListAttrAppendOnClient sends MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrAppendOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, val interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendData(val)
	return gwc.SendPacketRelease(packet)
}

// SendCallEntityMethodOnClient sends MT_CALL_ENTITY_METHOD_ON_CLIENT message
func (gwc *GoWorldConnection) SendCallEntityMethodOnClient(gateid uint16, clientid common.ClientID, entityID common.EntityID, method string, args []interface{}) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	return gwc.SendPacketRelease(packet)
}

// SendSetClientFilterProp sends MT_SET_CLIENTPROXY_FILTER_PROP message
func (gwc *GoWorldConnection) SendSetClientFilterProp(gateid uint16, clientid common.ClientID, key, val string) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_CLIENTPROXY_FILTER_PROP)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	return gwc.SendPacketRelease(packet)
}

// SendClearClientFilterProp sends MT_CLEAR_CLIENTPROXY_FILTER_PROPS message
func (gwc *GoWorldConnection) SendClearClientFilterProp(gateid uint16, clientid common.ClientID) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CLEAR_CLIENTPROXY_FILTER_PROPS)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	return gwc.SendPacketRelease(packet)
}

// SendCallFilterClientProxies sends MT_CALL_FILTERED_CLIENTS message
func (gwc *GoWorldConnection) SendCallFilterClientProxies(op FilterClientsOpType, key, val string, method string, args []interface{}) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_FILTERED_CLIENTS)
	packet.AppendByte(byte(op))
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	return gwc.SendPacketRelease(packet)
}

// SendQuerySpaceGameIDForMigrate sends MT_QUERY_SPACE_GAMEID_FOR_MIGRATE message
func (gwc *GoWorldConnection) SendQuerySpaceGameIDForMigrate(spaceid common.EntityID, entityid common.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_QUERY_SPACE_GAMEID_FOR_MIGRATE)
	packet.AppendEntityID(spaceid)
	packet.AppendEntityID(entityid)
	return gwc.SendPacketRelease(packet)
}

// SendMigrateRequest sends MT_MIGRATE_REQUEST message
func (gwc *GoWorldConnection) SendMigrateRequest(entityID common.EntityID, spaceID common.EntityID, spaceGameID uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_MIGRATE_REQUEST)
	packet.AppendEntityID(entityID)
	packet.AppendEntityID(spaceID)
	packet.AppendUint16(spaceGameID)
	return gwc.SendPacketRelease(packet)
}

// SendCancelMigrate sends MT_CANCEL_MIGRATE message to dispatcher to unblock the entity
func (gwc *GoWorldConnection) SendCancelMigrate(entityid common.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CANCEL_MIGRATE)
	packet.AppendEntityID(entityid)
	return gwc.SendPacketRelease(packet)
}

// SendRealMigrate sends MT_REAL_MIGRATE message
func (gwc *GoWorldConnection) SendRealMigrate(eid common.EntityID, targetGame uint16, targetSpace common.EntityID, x, y, z float32,
	typeName string, migrateData map[string]interface{}, timerData []byte, clientid common.ClientID, clientsrv uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_REAL_MIGRATE)
	packet.AppendEntityID(eid)
	packet.AppendUint16(targetGame)

	if !clientid.IsNil() {
		packet.AppendBool(true)
		packet.AppendClientID(clientid)
		packet.AppendUint16(clientsrv)
	} else {
		packet.AppendBool(false)
	}

	packet.AppendEntityID(targetSpace)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	packet.AppendVarStr(typeName)
	packet.AppendData(migrateData)
	packet.AppendVarBytes(timerData)

	return gwc.SendPacketRelease(packet)
}

// SendStartFreezeGame sends MT_START_FREEZE_GAME message
func (gwc *GoWorldConnection) SendStartFreezeGame(gameid uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_START_FREEZE_GAME)
	return gwc.SendPacketRelease(packet)
}

func MakeNotifyGameConnectedPacket(gameid uint16) *netutil.Packet {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(MT_NOTIFY_GAME_CONNECTED)
	pkt.AppendUint16(gameid)
	return pkt
}

func (gwc *GoWorldConnection) SendSetGameIDAck(dispid uint16, connectedGameIDs []uint16, rejectEntities []common.EntityID, srvdisRegisterMap map[string]string) error {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(MT_SET_GAME_ID_ACK)
	pkt.AppendUint16(dispid)
	pkt.AppendUint16(uint16(len(connectedGameIDs)))
	for _, gameid := range connectedGameIDs {
		pkt.AppendUint16(gameid)
	}
	// put rejected entity IDs to the packet
	pkt.AppendUint32(uint32(len(rejectEntities)))
	for _, eid := range rejectEntities {
		pkt.AppendEntityID(eid)
	}
	// put all services to the packet
	pkt.AppendMapStringString(srvdisRegisterMap)
	return gwc.SendPacketRelease(pkt)
}

// SendPacket send a packet to remote
func (gwc *GoWorldConnection) SendPacket(packet *netutil.Packet) error {
	return gwc.packetConn.SendPacket(packet)
}

// SendPacketRelease send a packet to remote and then release the packet
func (gwc *GoWorldConnection) SendPacketRelease(packet *netutil.Packet) error {
	err := gwc.packetConn.SendPacket(packet)
	packet.Release()
	return err
}

// Flush connection writes
func (gwc *GoWorldConnection) Flush(reason string) error {
	return gwc.packetConn.Flush(reason)
}

// SetAutoFlush starts a goroutine to flush connection writes at some specified interval
func (gwc *GoWorldConnection) SetAutoFlush(interval time.Duration) {
	if gwc.autoFlushing {
		gwlog.Panicf("%s.SetAutoFlush: already auto flushing!", gwc)
	}
	gwc.autoFlushing = true
	go func() {
		//defer gwlog.Debugf("%s: auto flush routine quited", gwc)
		for !gwc.IsClosed() {
			time.Sleep(interval)
			err := gwc.Flush("AutoFlush")
			if err != nil {
				break
			}
		}
	}()
}

// Recv receives the next packet and retrive the message type
func (gwc *GoWorldConnection) Recv(msgtype *MsgType) (*netutil.Packet, error) {
	pkt, err := gwc.packetConn.RecvPacket()
	if err != nil {
		return nil, err
	}

	*msgtype = MsgType(pkt.ReadUint16())
	if consts.DEBUG_PACKETS {
		gwlog.Infof("%s: Recv msgtype=%v, payload size=%d", gwc, *msgtype, pkt.GetPayloadLen())
	}
	return pkt, nil
}

// SetRecvDeadline set receive deadline
func (gwc *GoWorldConnection) SetRecvDeadline(deadline time.Time) error {
	return gwc.packetConn.SetRecvDeadline(deadline)
}

// Close this connection
func (gwc *GoWorldConnection) Close() error {
	gwc.closed.Store(true)
	return gwc.packetConn.Close()
}

// IsClosed returns if the connection is closed
func (gwc *GoWorldConnection) IsClosed() bool {
	return gwc.closed.Load()
}

// RemoteAddr returns the remote address
func (gwc *GoWorldConnection) RemoteAddr() net.Addr {
	return gwc.packetConn.RemoteAddr()
}

// LocalAddr returns the local address
func (gwc *GoWorldConnection) LocalAddr() net.Addr {
	return gwc.packetConn.LocalAddr()
}

func (gwc *GoWorldConnection) String() string {
	return gwc.packetConn.String()
}
