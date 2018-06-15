# GoWorld
_**Scalable Distributed Game Server Engine with Hot Reload in Golang**_

[![GoDoc](https://godoc.org/github.com/xiaonanln/goworld?status.png)](https://godoc.org/github.com/xiaonanln/goworld) [![Build Status](https://api.travis-ci.org/xiaonanln/goworld.svg?branch=master)](https://travis-ci.org/xiaonanln/goworld) [![Go Report Card](https://goreportcard.com/badge/github.com/xiaonanln/goworld)](https://goreportcard.com/report/github.com/xiaonanln/goworld) [![codecov](https://codecov.io/gh/xiaonanln/goworld/branch/master/graph/badge.svg)](https://codecov.io/gh/xiaonanln/goworld) [![ApacheLicense](https://img.shields.io/badge/license-APACHE%20License-blue.svg)](https://raw.githubusercontent.com/xiaonanln/goworld/master/LICENSE)
[![Join the chat at https://gitter.im/goworldgs/Lobby](https://badges.gitter.im/goworldgs/Lobby.svg)](https://gitter.im/goworldgs/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

---------------------------------------
  * [Features](#features)
  * [Architecture](#architecture)
  * [Introduction](#introduction)
  * [Get GoWorld](#get-goworld)
  * [Manage GoWorld Servers](#manage-goworld-servers)
  * [Demos](#demos)
    * [Chatroom Demo](#chatroom-demo)
    * [Unity Demo](#unity-demo)
---------------------------------------
>  中文资料：
> [游戏服务器介绍](http://www.cnblogs.com/isaiah/p/7259036.html)
> [目录结构说明](https://github.com/xiaonanln/goworld/wiki/GoWorld%E6%B8%B8%E6%88%8F%E6%9C%8D%E5%8A%A1%E5%99%A8%E5%BC%95%E6%93%8E%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84) 
> [使用GoWorld轻松实现分布式聊天服务器](https://github.com/xiaonanln/goworld/wiki/%E4%BD%BF%E7%94%A8GoWorld%E6%B8%B8%E6%88%8F%E6%9C%8D%E5%8A%A1%E5%99%A8%E5%BC%95%E6%93%8E%E8%BD%BB%E6%9D%BE%E5%AE%9E%E7%8E%B0%E5%88%86%E5%B8%83%E5%BC%8F%E8%81%8A%E5%A4%A9%E6%9C%8D%E5%8A%A1%E5%99%A8)
> **欢迎加入Go服务端开发交流群：[662182346](http://shang.qq.com/wpa/qunwpa?idkey=f2a99bd9bd9e6df3528174180aad753d05b372a8828e1b8e5c1ec5df42b301db)**
---------------------------------------  
## Features
* **Spaces & Entities**: manage multiple spaces and entities with AOI support
* **Distributed**: increase server capacity by using more machines
* **Hot-Swappable**: update game logic by restarting server process
* **Multiple Communication Protocols**: supports TCP, [KCP](https://github.com/skywind3000/kcp) and WebSocket
* **Traffic Compression & Encryption**: traffic between clients and servers can be compressed and encrypted

## Architecture
![GoWorld Architecture](http://goworldgs.com/public/img/goworld_arch.png "GoWorld Architecture")

## Introduction
GoWorld server adopts an entity framework, in which entities represent all players, monsters, NPCs.
Entities in the same space can visit each other directly by calling methods or access attributes. 
Entities in different spaces can call each over using RPC.

A GoWorld server consists of one dispatcher, one or more games and one or more gates. 
The gates are responsible for handling client connections and receive/send packets from/to clients. 
The games manages all entities and runs all game logic. 
The dispatcher is responsible for redirecting packets among games and between games and gates.  

The game processes are **hot-swappable**. 
We can swap a game by sending SIGUSR1 to the process and restart the process with **-restore** parameter to bring game 
back to work but with the latest executable image. This feature enables updating server-side logic or fixing server bugs
 transparently without significant interference of online players. 

## Get GoWorld
```bash
go get github.com/xiaonanln/goworld/cmd/...
```

### Install dependencies

GoWorld uses [Dep](https://github.com/golang/dep) to manage packages. 

Install Go [Dep](https://github.com/golang/dep#setup):
```bash
go get -u github.com/golang/dep/cmd/dep
```
Install Dependencies: 
```go
cd $GOPATH/src/github.com/xiaonanln/goworld
dep ensure
```

**For Chinese only:** 
中国的朋友可能会碰到网络问题，那么您可以通过点击下载[vendor.tar.gz](http://goworldgs.com/goworld_files/vendor.tar.gz)，并将其解压到goworld根目录下。

## Manage GoWorld Servers
Use command `goworld` to build, start, stop and reload game servers. 

**Build Example Chatroom Server:**
```bash
goworld build examples/chatroom_demo
```

**Start Example Chatroom Server: (dispatcher -> game -> gate)**
```bash
goworld start examples/chatroom_demo
``` 

**Stop Game Server (gate -> game -> dispatcher):**
```bash
goworld stop examples/chatroom_demo
```

**Reload Game Servers:**
```bash
goworld reload examples/chatroom_demo
```
Reload will reboot game processes with the current executable while preserving all game server states. 
**However, it is not workable on Windows.**  


## Demos

### Chatroom Demo
The chatroom demo is a simple implementation of chatroom server and client. It illustrates how
GoWorld can also be used for games which don't divide players by spaces. 

The chatroom demo provides following features:
* register & login
* send chat message
* switch chatrooms

**Build Server:**
```bash
goworld build examples/chatroom_demo
```
**Run Server:**
```bash
goworld start examples/chatroom_demo
```

**Chatroom Demo Client:**

Chatroom demo client implements the client-server protocol in Javascript.  
The client for chatroom demo is hosted at [github.com/xiaonanln/goworld-chatroom-demo-client](https://github.com/xiaonanln/goworld-chatroom-demo-client).
The project was created and built in [Cocos Creater 1.5](http://www.cocos2d-x.org/). A running server and client demo can also be found at [http://goworldgs.com/chatclient/](http://goworldgs.com/chatclient/).

### Unity Demo
Unity Demo is a simple multiple player monster shooting game to show how spaces and entities of GoWorld
can be used to create multiple player online games.  

* register & login
* move players in space
* summon monsters
* player shoot monsters
* monsters attack players

Developing features:
* Hit effects
* Players migrate among multiple spaces
* Server side map navigation

**Build Server:**
```bash
goworld build examples/unity_demo
```
**Run Server:**
```bash
goworld start examples/unity_demo
```

**Unity Demo Client:**

Unity demo client implements the client-server protocol in C#. 
The client for unity demo is hosted at [https://github.com/xiaonanln/goworld-unity-demo](https://github.com/xiaonanln/goworld-unity-demo).
The project was created and built in [Unity 2017.1](https://unity3d.com/). 

You can try the demo by downloading [GoWorldUnityDemo.zip](https://pan.baidu.com/s/1MSs8cUAXUnQPTUCN-lIpHg). 
The demo connects to a goworld server on Amazon EC2 instance. 