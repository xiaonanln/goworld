# v0.1
* ListAttr
* Support MySQL for kvdb and entity storage
* Connection Encryption
* Logging system improvement
* Chatroom Demo with Cocos2d-x Client?

# Optimizations
* Freezing game might take a lot memory, need optimization

# Enhancing Fault-tolerance
* make sure entity is destroyed if creation fails

# MISC
* Not allow client connection if server is not ready
* seperate functionalities of PacketConnection and BufferedConnection
* Re-implement Service Architecture: auto creation, auto dispatch

# Subsystem to be implemented
* Client system on Unity3D ?
* more documents
* Plugin system ?

# Future plans
* use protobuf
* Session based Connection Management
* Entity Client Reconnect Mechianism
* use vendor for submodules
* use termui for server status monitoring
* use UDP : e.x. KCP
* What is websocket?
* The state machine server programming mechanism