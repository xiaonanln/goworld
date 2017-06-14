package entity

import (
	"fmt"
	"reflect"

	"time"

	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/typeconv"

	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage"
)

type Entity struct {
	ID       EntityID
	TypeName string
	I        IEntity
	IV       reflect.Value

	destroyed  bool
	rpcDescMap RpcDescMap
	space      *Space
	aoi        AOI
	timers     map[*timer.Timer]struct{}
	client     *GameClient
}

// Functions declared by IEntity can be override in Entity subclasses
type IEntity interface {
	// Entity Lifetime
	OnInit()
	OnCreated()
	OnDestroy()
	// Space Operations
	OnEnterSpace()
	OnLeaveSpace(space *Space)
	// Storage: Save & Load
	IsPersistent() bool
	GetPersistentData() map[string]interface{}
	LoadPersistentData(data map[string]interface{})
	// Client Notifications
	OnClientConnected()
	OnClientDisconnected()
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *Entity) Destroy() {
	if e.destroyed {
		return
	}

	gwlog.Info("%s.Destroy.", e)
	if e.space != nil {
		e.space.leave(e)
	}
	e.clearTimers()
	e.I.OnDestroy()
	entityManager.del(e.ID)
	e.destroyed = true
}

func (e *Entity) IsDestroyed() bool {
	return e.destroyed
}

func (e *Entity) Save() {
	if !e.I.IsPersistent() {
		return
	}

	if consts.DEBUG_SAVE_LOAD {
		gwlog.Debug("SAVING %s ...", e)
	}

	data := e.I.GetPersistentData()

	storage.Save(e.TypeName, e.ID, data)
}

func (e *Entity) setupSaveTimer() {
	e.AddTimer(consts.SAVE_INTERVAL, e.Save)
}

// Space Operations related to entity

// Interests and Uninterest among entities
func (e *Entity) interest(other *Entity) {
	e.aoi.interest(other)
}

func (e *Entity) uninterest(other *Entity) {
	e.aoi.uninterest(other)
}

func (e *Entity) Neighbors() EntitySet {
	return e.aoi.neighbors
}

// Timer & Callback Management
func (e *Entity) AddCallback(d time.Duration, cb timer.CallbackFunc) {
	var t *timer.Timer
	t = timer.AddCallback(d, func() {
		delete(e.timers, t)
		cb()
	})
	e.timers[t] = struct{}{}
}

func (e *Entity) Post(cb timer.CallbackFunc) {
	e.AddCallback(0, cb)
}

func (e *Entity) AddTimer(d time.Duration, cb timer.CallbackFunc) {
	t := timer.AddTimer(d, cb)
	e.timers[t] = struct{}{}
}

func (e *Entity) clearTimers() {
	for t := range e.timers {
		t.Cancel()
	}
	e.timers = map[*timer.Timer]struct{}{}
}

// Call other entities
func (e *Entity) Call(id EntityID, method string, args ...interface{}) {
	callRemote(id, method, args)
}

func (e *Entity) onCall(methodName string, args []interface{}) {
	defer func() {
		err := recover() // recover from any error during RPC call
		if err != nil {
			gwlog.TraceError("%s.%s%v paniced: %s", e, methodName, args, err)
		}
	}()

	rpcDesc := e.rpcDescMap[methodName]
	methodType := rpcDesc.MethodType

	if rpcDesc.NumArgs != len(args) {
		gwlog.Error("Method %s receives %d arguments, but given %d: %v", methodName, rpcDesc.NumArgs, len(args), args)
		return
	}

	in := make([]reflect.Value, len(args)+1)
	in[0] = reflect.ValueOf(e.I)
	for i, arg := range args {
		argType := methodType.In(i + 1)
		in[i+1] = typeconv.Convert(arg, argType)
	}
	rpcDesc.Func.Call(in)
}

// Register for global service
func (e *Entity) DeclareService(serviceName string) {
	dispatcher_client.GetDispatcherClientForSend().SendDeclareService(e.ID, serviceName)
}

// Default Handlers
func (e *Entity) OnInit() {
	gwlog.Warn("%s.OnInit not implemented", e)
}

func (e *Entity) OnCreated() {
	gwlog.Debug("%s.OnCreated", e)
}

func (e *Entity) OnEnterSpace() {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.OnEnterSpace >>> %s", e, e.space)
	}
}

func (e *Entity) OnLeaveSpace(space *Space) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.OnLeaveSpace <<< %s", e, space)
	}
}

func (e *Entity) OnDestroy() {
}

// Default handlers for persistence

func (e *Entity) IsPersistent() bool {
	return false
}

func (e *Entity) GetPersistentData() map[string]interface{} {
	gwlog.TraceError("%s.GetPersistentData not implemented", e)
	return nil
}

func (e *Entity) LoadPersistentData(data map[string]interface{}) {
	gwlog.TraceError("%s.LoadPersistentData not implemented", e)
}

// Clients
func (e *Entity) GetClient() *GameClient {
	return e.client
}

func (e *Entity) SetClient(client *GameClient) {
	oldClient := e.client
	if oldClient == client {
		return
	}

	e.client = client
	if oldClient != nil {
		// send destroy entity to client
		client.SendDestroyEntity(e)
	}

	if client != nil {
		// send create entity to client
		client.SendCreateEntity(e)
	}

	if oldClient == nil && client != nil {
		// got net client
		e.I.OnClientConnected()
	} else if oldClient != nil && client == nil {
		e.I.OnClientDisconnected()
	}
}
func (e *Entity) OnClientConnected() {
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.OnClientConnected: %s", e, e.client)
	}
}

func (e *Entity) OnClientDisconnected() {
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.OnClientDisconnected: %s", e, e.client)
	}
}
