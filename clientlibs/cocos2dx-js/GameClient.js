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
        this.connect()
    },

    // called every frame, uncomment this function to activate update callback
    update: function (dt) {

    },
    
    connect: function() {
        var serverAddr = 'ws://'+this.serverAddr+':'+this.serverPort+'/ws'
        console.log("正在连接 " + serverAddr + ' ...')
        var websocket = new WebSocket(serverAddr)
        websocket.binaryType = 'blob'
        console.log(websocket)
        
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
               console.log("收到数据：", typeof(event.data), event.data.length,  event.data.size);
               for (var key in event) {
                console.log(key, event[key])
               }
          }
    
           //连接关闭的回调方法
           websocket.onclose = function () {
              console.log("WebSocket连接关闭");
           }
    
           //监听窗口关闭事件，当窗口关闭时，主动去关闭websocket连接，防止连接还没断开就关闭窗口，server端会抛异常。
           window.onbeforeunload = function () {
               console.log("onbeforeunload");
           }
    }
});
