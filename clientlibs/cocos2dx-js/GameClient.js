var msgpack = require("msgpack");
var ClientEntity = require("ClientEntity")

const _RECV_PAYLOAD_LENGTH = 1
const _RECV_PAYLOAD = 2

const CLIENTID_LENGTH = 16
const ENTITYID_LENGTH = 16

const MT_GATE_SERVICE_MSG_TYPE_START = 1000
const MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START = 1001 // messages that should be redirected to client proxy
const MT_CREATE_ENTITY_ON_CLIENT = 1002
const MT_DESTROY_ENTITY_ON_CLIENT = 1003
const MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT = 1004
const MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT = 1005
const MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT = 1006
const MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT = 1007
const MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT = 1008
const MT_CALL_ENTITY_METHOD_ON_CLIENT = 1009
const MT_UPDATE_POSITION_ON_CLIENT = 1010
const MT_UPDATE_YAW_ON_CLIENT = 1011
const MT_SET_CLIENTPROXY_FILTER_PROP = 1012
const MT_CLEAR_CLIENTPROXY_FILTER_PROPS = 1013

// add more ...

const MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP = 1500

const MT_CALL_FILTERED_CLIENTS = 1501
const MT_SYNC_POSITION_YAW_ON_CLIENTS = 1502

// add more ...

const MT_GATE_SERVICE_MSG_TYPE_STOP = 2000


cc.Class({
    extends: cc.Component,

    properties: {
        // foo: {
        //    default: null,      // The default value will be used only when the component attaching
        //                           to a node for the first time
        //    url: cc.Texture2D,  // optional, default is typeof default
        //    serializable: true, // optional, default is true
        //    visible: true,      // optional, default is true
        //    displayName: 'Foo', // optional
        //    readonly: false,    // optional, default is false
        // },
        // ...
        serverAddr: {
            default: '',
        },
        serverPort: {
            default: '',
        },
    },

    // use this for initialization
    onLoad: function () {
        this.recvBuf = new ArrayBuffer()
        this.recvStatus = _RECV_PAYLOAD_LENGTH
        this.recvPayloadLen = 0
        this.entities = {}
        this.connect()
    },

    // called every frame, uncomment this function to activate update callback
    update: function (dt) {

    },

    onRecvData: function (data) {
        if (this.recvBuf.byteLength == 0) {
            this.recvBuf = data
        } else {
            var tmp = new Uint8Array( this.recvBuf.byteLength + data.byteLength );
            tmp.set( new Uint8Array( this.recvBuf ), 0 );
            tmp.set( new Uint8Array( data ), this.recvBuf.byteLength );
            this.recvBuf = tmp.buffer
        }

        console.log("未处理数据：", this.recvBuf.byteLength)
        let payload = this.tryReceivePacket()
        if (payload !== null) {
            this.onReceivePacket(payload)
        }
    },

    // 从已经收到的数据（recvBuf）里解析出数据包（Packet）
    tryReceivePacket: function() {
        if (this.recvStatus == _RECV_PAYLOAD_LENGTH) {
            if (this.recvBuf.byteLength < 4) {
                return null
            }
            [this.recvPayloadLen, this.recvBuf] = this.readUint32(this.recvBuf)
            console.log("数据包大小: ", this.recvPayloadLen)
            this.recvStatus = _RECV_PAYLOAD
        }

        // recv status == _RECV_PAYLOAD
        console.log("包大小：", this.recvPayloadLen, "现有数据：", this.recvBuf.byteLength)
        if (this.recvBuf.byteLength < this.recvPayloadLen) {
            // payload not enough
            return null
        }

        // 足够了，返回包数据
        var payload = this.recvBuf.slice(0, this.recvPayloadLen)
        this.recvBuf = this.recvBuf.slice(this.recvPayloadLen)
        // 恢复到接收长度状态
        this.recvStatus = _RECV_PAYLOAD_LENGTH
        this.recvPayloadLen = 0
        return payload
    },

    onReceivePacket: function (payload) {
            var [msgtype, payload] = this.readUint16(payload)
            console.log("收到包：", payload, payload.byteLength, "，消息类型：", msgtype)
            if (msgtype != MT_CALL_FILTERED_CLIENTS && msgtype != MT_SYNC_POSITION_YAW_ON_CLIENTS) {
                var [dummy, payload] = this.readUint16(payload)
                var [dummy, payload] = this.readBytes(payload, CLIENTID_LENGTH) // read ClientID
            }

            if (msgtype == MT_CREATE_ENTITY_ON_CLIENT) {
                var [isPlayer, payload] = this.readBool(payload)
                var [eid, payload] = this.readEntityID(payload)
                var [typeName, payload] = this.readVarStr(payload)
                var [x, payload] = this.readFloat32(payload)
                var [y, payload] = this.readFloat32(payload)
                var [z, payload] = this.readFloat32(payload)
                var [yaw, payload] = this.readFloat32(payload)
                var [clientData,payload] = this.readVarBytes(payload)
                clientData = msgpack.decode(clientData)
                console.log("MT_CREATE_ENTITY_ON_CLIENT", "isPlayer", isPlayer, 'eid', eid,"typeName", typeName, 'position', x, y, z, 'yaw', yaw, 'clientData', JSON.stringify(clientData))
                
                var e = new ClientEntity()
                e.create( typeName, eid )
                this.entities[eid] = e
                this.onEntityCreated(e)
                e.onCreated()
            }
    },

    readUint8: function(buf) {
        let v = new Uint8Array(buf)[0]
        return [v, buf.slice(1)]
    },
    readUint16: function(buf) {
        let v = new Uint16Array(buf.slice(0, 2))[0]
        return [v, buf.slice(2)]
    },
    readUint32: function(buf) {
        let v = new Uint32Array(buf.slice(0, 4))[0]
        return [v, buf.slice(4)]
    },
    readFloat32: function(buf) {
        let v = new Float32Array(buf.slice(0, 4))[0]
        return [v, buf.slice(4)]
    },
    readBytes: function(buf, length) {
        let v = new Uint8Array(buf.slice(0, length))
        return [v, buf.slice(length)]
    },
    readVarBytes: function(buf) {
        var [n, buf] = this.readUint32(buf)
        var [b, buf] = this.readBytes(buf, n)
        return [b, buf]
    },
    readEntityID: function(buf) {
        var [eid, buf] = this.readBytes(buf, ENTITYID_LENGTH)
        eid = String.fromCharCode.apply(null, eid)
        return [eid, buf]
    },
    readVarStr: function(buf) {
        var [b, buf] = this.readVarBytes(buf)
        let s = String.fromCharCode.apply(null, b)
        return [s, buf]
    },
    readBool: function(buf) {
        var b
        [b, buf] = this.readUint8(buf)
        b = b == 0 ? false : true
        return [b, buf]
    },

    connect: function() {
        var serverAddr = 'ws://'+this.serverAddr+':'+this.serverPort+'/ws'
        console.log("正在连接 " + serverAddr + ' ...')
        var websocket = new WebSocket(serverAddr)
        websocket.binaryType = 'blob'
        console.log(websocket)
        var gameclient = this

          //连接发生错误的回调方法
          websocket.onerror = function () {
               console.log("WebSocket连接发生错误");
          };

           //连接成功建立的回调方法
           websocket.onopen = function () {
               console.log("WebSocket连接成功");
           }

          //接收到消息的回调方法
           websocket.onmessage = function (event) {
               var data = event.data
               console.log("收到数据：", typeof(data), data.length);
               gameclient.onRecvData(data)
          }

           //连接关闭的回调方法
           websocket.onclose = function () {
              console.log("WebSocket连接关闭");
           }

           //监听窗口关闭事件，当窗口关闭时，主动去关闭websocket连接，防止连接还没断开就关闭窗口，server端会抛异常。
           window.onbeforeunload = function () {
               console.log("onbeforeunload");
           }
    },
    
    onEntityCreated: function(e) {
        console.log("entity created:", e.toString())
    }
    
});
