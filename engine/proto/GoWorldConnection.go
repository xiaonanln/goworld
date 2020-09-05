package proto

import (
	"github.com/xiaonanln/pktconn"
	"net"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/netutil"
)

// GoWorldConnection is the network protocol implementation of GoWorld components (dispatcher, gate, game)
type GoWorldConnection struct {
	packetConn   *netutil.PacketConnection
	closed       xnsyncutil.AtomicBool
	autoFlushing bool
}

// NewGoWorldConnection creates a GoWorldConnection using network connection
func NewGoWorldConnection(conn netutil.Connection, tag interface{}) *GoWorldConnection {
	return &GoWorldConnection{
		packetConn: netutil.NewPacketConnection(conn, tag),
	}
}

// SendSetGameID sends MT_SET_GAME_ID message
func (gwc *GoWorldConnection) SendSetGameID(id uint16, isReconnect bool, isRestore bool, isBanBootEntity bool,
	eids []common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GAME_ID)
	packet.AppendUint16(id)
	packet.AppendBool(isReconnect)
	packet.AppendBool(isRestore)
	packet.AppendBool(isBanBootEntity)
	// put all entity IDs to the packet

	packet.AppendUint32(uint32(len(eids)))
	for _, eid := range eids {
		packet.AppendEntityID(eid)
	}
	gwc.SendPacketRelease(packet)
}

// SendSetGateID sends MT_SET_GATE_ID message
func (gwc *GoWorldConnection) SendSetGateID(id uint16) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GATE_ID)
	packet.AppendUint16(id)
	gwc.SendPacketRelease(packet)
}

// SendNotifyCreateEntity sends MT_NOTIFY_CREATE_ENTITY message
func (gwc *GoWorldConnection) SendNotifyCreateEntity(id common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CREATE_ENTITY)
	packet.AppendEntityID(id)
	gwc.SendPacketRelease(packet)
}

// SendNotifyDestroyEntity sends MT_NOTIFY_DESTROY_ENTITY message
func (gwc *GoWorldConnection) SendNotifyDestroyEntity(id common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_DESTROY_ENTITY)
	packet.AppendEntityID(id)
	gwc.SendPacketRelease(packet)
}

// SendNotifyClientConnected sends MT_NOTIFY_CLIENT_CONNECTED message
func (gwc *GoWorldConnection) SendNotifyClientConnected(id common.ClientID, bootEid common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_CONNECTED)
	packet.AppendClientID(id)
	packet.AppendEntityID(bootEid)
	gwc.SendPacketRelease(packet)
}

// SendNotifyClientDisconnected sends MT_NOTIFY_CLIENT_DISCONNECTED message
func (gwc *GoWorldConnection) SendNotifyClientDisconnected(id common.ClientID, ownerEntityID common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_DISCONNECTED)
	packet.AppendEntityID(ownerEntityID)
	packet.AppendClientID(id)
	gwc.SendPacketRelease(packet)
}

// SendCreateEntitySomewhere sends MT_CREATE_ENTITY_SOMEWHERE message
func (gwc *GoWorldConnection) SendCreateEntitySomewhere(gameid uint16, entityid common.EntityID, typeName string, data map[string]interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_SOMEWHERE)
	packet.AppendUint16(gameid)
	packet.AppendEntityID(entityid)
	packet.AppendVarStr(typeName)
	packet.AppendData(data)
	gwc.SendPacketRelease(packet)
}

// SendLoadEntitySomewhere sends MT_LOAD_ENTITY_SOMEWHERE message
func (gwc *GoWorldConnection) SendLoadEntitySomewhere(typeName string, entityID common.EntityID, gameid uint16) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_LOAD_ENTITY_SOMEWHERE)
	packet.AppendUint16(gameid)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(typeName)
	gwc.SendPacketRelease(packet)
}

// SendKvregRegister
func (gwc *GoWorldConnection) SendKvregRegister(srvid string, info string, force bool) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_KVREG_REGISTER)
	packet.AppendVarStr(srvid)
	packet.AppendVarStr(info)
	packet.AppendBool(force)
	gwc.SendPacketRelease(packet)
}

// SendCallEntityMethod sends MT_CALL_ENTITY_METHOD message
func (gwc *GoWorldConnection) SendCallEntityMethod(id common.EntityID, method string, args []interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	gwc.SendPacketRelease(packet)
}

// SendCallEntityMethodFromClient sends MT_CALL_ENTITY_METHOD_FROM_CLIENT message
func (gwc *GoWorldConnection) SendCallEntityMethodFromClient(id common.EntityID, method string, args []interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_FROM_CLIENT)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	gwc.SendPacketRelease(packet)
}

// SendCreateEntityOnClient sends MT_CREATE_ENTITY_ON_CLIENT message
func (gwc *GoWorldConnection) SendCreateEntityOnClient(gameid uint16, clientid common.ClientID, typeName string, entityid common.EntityID,
	isPlayer bool, clientData map[string]interface{}, x, y, z float32, yaw float32) {
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
	gwc.SendPacketRelease(packet)
}

// SendSyncPositionYawFromClient sends MT_SYNC_POSITION_YAW_FROM_CLIENT message
func (gwc *GoWorldConnection) SendSyncPositionYawFromClient(entityID common.EntityID, x, y, z float32, yaw float32) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SYNC_POSITION_YAW_FROM_CLIENT)
	packet.AppendEntityID(entityID)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	packet.AppendFloat32(yaw)
	gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SetHeartbeatFromClient() {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_HEARTBEAT_FROM_CLIENT)
	gwc.SendPacketRelease(packet)
}

// SendDestroyEntityOnClient sends MT_DESTROY_ENTITY_ON_CLIENT message
func (gwc *GoWorldConnection) SendDestroyEntityOnClient(gateid uint16, clientid common.ClientID, typeName string, entityid common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DESTROY_ENTITY_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(typeName)
	packet.AppendEntityID(entityid)
	gwc.SendPacketRelease(packet)
}

// SendNotifyMapAttrChangeOnClient sends MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrChangeOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, key string, val interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendVarStr(key)
	packet.AppendData(val)
	gwc.SendPacketRelease(packet)
}

// SendNotifyMapAttrDelOnClient sends MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrDelOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, key string) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendVarStr(key)
	gwc.SendPacketRelease(packet)
}

// SendNotifyMapAttrClearOnClient sends MT_NOTIFY_MAP_ATTR_CLEAR_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrClearOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_CLEAR_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	gwc.SendPacketRelease(packet)
}

// SendNotifyListAttrChangeOnClient sends MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrChangeOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, index uint32, val interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendUint32(index)
	packet.AppendData(val)
	gwc.SendPacketRelease(packet)
}

// SendNotifyListAttrPopOnClient sends MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrPopOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	gwc.SendPacketRelease(packet)
}

// SendNotifyListAttrAppendOnClient sends MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrAppendOnClient(gateid uint16, clientid common.ClientID, entityid common.EntityID, path []interface{}, val interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendData(val)
	gwc.SendPacketRelease(packet)
}

// SendCallEntityMethodOnClient sends MT_CALL_ENTITY_METHOD_ON_CLIENT message
func (gwc *GoWorldConnection) SendCallEntityMethodOnClient(gateid uint16, clientid common.ClientID, entityID common.EntityID, method string, args []interface{}) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_ON_CLIENT)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	gwc.SendPacketRelease(packet)
}

// SendSetClientFilterProp sends MT_SET_CLIENTPROXY_FILTER_PROP message
func (gwc *GoWorldConnection) SendSetClientFilterProp(gateid uint16, clientid common.ClientID, key, val string) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_CLIENTPROXY_FILTER_PROP)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	gwc.SendPacketRelease(packet)
}

// SendClearClientFilterProp sends MT_CLEAR_CLIENTPROXY_FILTER_PROPS message
func (gwc *GoWorldConnection) SendClearClientFilterProp(gateid uint16, clientid common.ClientID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CLEAR_CLIENTPROXY_FILTER_PROPS)
	packet.AppendUint16(gateid)
	packet.AppendClientID(clientid)
	gwc.SendPacketRelease(packet)
}

// SendCallFilterClientProxies sends MT_CALL_FILTERED_CLIENTS message
func AllocCallFilterClientProxiesPacket(op FilterClientsOpType, key, val string, method string, args []interface{}) *netutil.Packet {
	packet := netutil.NewPacket()
	packet.AppendUint16(MT_CALL_FILTERED_CLIENTS)
	packet.AppendByte(byte(op))
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	return packet
}

func AllocCallNilSpacesPacket(exceptGameID uint16, method string, args []interface{}) *netutil.Packet {
	// construct one packet for multiple sending
	packet := netutil.NewPacket()
	packet.AppendUint16(MT_CALL_NIL_SPACES)
	packet.AppendUint16(exceptGameID)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	return packet
}

func AllocGameLBCInfoPacket(lbcinfo GameLBCInfo) *netutil.Packet {
	packet := netutil.NewPacket()
	packet.AppendUint16(MT_GAME_LBC_INFO)
	packet.AppendData(lbcinfo)
	return packet
}

// SendQuerySpaceGameIDForMigrate sends MT_QUERY_SPACE_GAMEID_FOR_MIGRATE message
func (gwc *GoWorldConnection) SendQuerySpaceGameIDForMigrate(spaceid common.EntityID, entityid common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_QUERY_SPACE_GAMEID_FOR_MIGRATE)
	packet.AppendEntityID(spaceid)
	packet.AppendEntityID(entityid)
	gwc.SendPacketRelease(packet)
}

// SendMigrateRequest sends MT_MIGRATE_REQUEST message
func (gwc *GoWorldConnection) SendMigrateRequest(entityID common.EntityID, spaceID common.EntityID, spaceGameID uint16) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_MIGRATE_REQUEST)
	packet.AppendEntityID(entityID)
	packet.AppendEntityID(spaceID)
	packet.AppendUint16(spaceGameID)
	gwc.SendPacketRelease(packet)
}

// SendCancelMigrate sends MT_CANCEL_MIGRATE message to dispatcher to unblock the entity
func (gwc *GoWorldConnection) SendCancelMigrate(entityid common.EntityID) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CANCEL_MIGRATE)
	packet.AppendEntityID(entityid)
	gwc.SendPacketRelease(packet)
}

// SendRealMigrate sends MT_REAL_MIGRATE message
func (gwc *GoWorldConnection) SendRealMigrate(eid common.EntityID, targetGame uint16, data []byte) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_REAL_MIGRATE)
	packet.AppendEntityID(eid)
	packet.AppendUint16(targetGame)
	packet.AppendVarBytes(data)
	gwc.SendPacketRelease(packet)
}

func AllocStartFreezeGamePacket() *netutil.Packet {
	packet := netutil.NewPacket()
	packet.AppendUint16(MT_START_FREEZE_GAME)
	return packet
}

func MakeNotifyGameConnectedPacket(gameid uint16) *netutil.Packet {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(MT_NOTIFY_GAME_CONNECTED)
	pkt.AppendUint16(gameid)
	return pkt
}

func MakeNotifyGameDisconnectedPacket(gameid uint16) *netutil.Packet {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(MT_NOTIFY_GAME_DISCONNECTED)
	pkt.AppendUint16(gameid)
	return pkt
}

func MakeNotifyDeploymentReadyPacket() *netutil.Packet {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(MT_NOTIFY_DEPLOYMENT_READY)
	return pkt
}

func (gwc *GoWorldConnection) SendSetGameIDAck(dispid uint16, isDeploymentReady bool, connectedGameIDs []uint16, rejectEntities []common.EntityID, kvregRegisterMap map[string]string) {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(MT_SET_GAME_ID_ACK)
	pkt.AppendUint16(dispid)

	pkt.AppendBool(isDeploymentReady)

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
	pkt.AppendMapStringString(kvregRegisterMap)
	gwc.SendPacketRelease(pkt)
}

// SendPacket send a packet to remote
func (gwc *GoWorldConnection) SendPacket(packet *netutil.Packet) {
	gwc.packetConn.SendPacket(packet)
}

// SendPacketRelease send a packet to remote and then release the packet
func (gwc *GoWorldConnection) SendPacketRelease(packet *netutil.Packet) {
	gwc.packetConn.SendPacket(packet)
	packet.Release()
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

func (gwc *GoWorldConnection) RecvChan(recvChan chan *pktconn.Packet) error {
	return gwc.packetConn.RecvChan(recvChan)
}
