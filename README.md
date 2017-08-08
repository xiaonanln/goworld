# GoWorld [![GoDoc](https://godoc.org/github.com/xiaonanln/goworld?status.png)](https://godoc.org/github.com/xiaonanln/goworld) [![Build Status](https://api.travis-ci.org/xiaonanln/goworld.svg?branch=master)](https://travis-ci.org/xiaonanln/goworld) [![Go Report Card](https://goreportcard.com/badge/github.com/xiaonanln/goworld)](https://goreportcard.com/report/github.com/xiaonanln/goworld)

_**Scalable Distributed Game Server Engine with Hot Reload in Golang**_
---------------------------------------
  * [Features](#features)
  * [Architecture](#architecture)
  * [Introduction](#introduction)
  * [Get GoWorld](#get-goworld)
  * [Run Example Server & Client](#run-example-server--client)
---------------------------------------
>  中文资料：
> [游戏服务器介绍](http://www.cnblogs.com/isaiah/p/7259036.html)
> [目录结构说明](https://github.com/xiaonanln/goworld/wiki/GoWorld%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84) 
> **欢迎加入Go服务端开发交流群：[662182346](http://shang.qq.com/wpa/qunwpa?idkey=f2a99bd9bd9e6df3528174180aad753d05b372a8828e1b8e5c1ec5df42b301db)**
---------------------------------------  
## Features
* **Spaces & Entities**: manage multiple spaces and entities with AOI support
* **Distributed**: increase server capacity by using more machines
* **Hot-Swappable**: update game logic by restarting server process

## Architecture
![GoWorld Architecture](http://goworldgs.com/static/goworld_arch.png "GoWorld Architecture")

## Introduction
GoWorld server adopts an entity framework, in which entities represent all players, monsters, NPCs.
Entities in the same space can visit each other directly by calling methods or access attributes. 
Entities in different spaces can call each over using RPC.

A GoWorld server consists of one dispatcher, one or more games and one or more gates. 
The gates are responsable for handling client connections and receive/send packets from/to clients. 
The games manages all entities and runs all game logic. 
The dispatcher is responsable for redirecting packets among games and between games and gates.  

The game processes are **hot-swappable**. 
We can swap a game by sending SIGUSR1 to the process and restart the process with **-restore** parameter to bring game 
back to work but with the latest executive image. This feature enables updating server-side logic or fixing server bugs
 transparently without significant interference of online players. 

## Get GoWorld
**Download goworld:**
```bash
get github.com/xiaonanln/goworld
```

**Install dependencies**
```bash
go get -u github.com/xiaonanln/go-xnsyncutil/xnsyncutil
go get -u github.com/xiaonanln/goTimer
go get -u github.com/xiaonanln/typeconv
go get -u golang.org/x/net/context
go get -u github.com/Sirupsen/logrus
go get -u github.com/garyburd/redigo/redis
go get -u github.com/google/btree
go get -u github.com/pkg/errors
go get -u golang.org/x/net/websocket
go get -u gopkg.in/eapache/queue.v1
go get -u gopkg.in/ini.v1
go get -u gopkg.in/mgo.v2
go get -u gopkg.in/vmihailenco/msgpack.v2
go get -u gopkg.in/natefinch/lumberjack.v2

```

## Run Example Server & Client
1. Get MongoDB or Redis Running
2. Copy goworld.ini.sample to goworld.ini, and configure accordingly
    ```bash
    cp goworld.ini.sample goworld.ini
    ```
3. Build and run dispatcher:
    ```bash
    make dispatcher
    components/dispatcher/dispatcher
    ```

4. Build and run gate:
    ```bash
    make gate
    components/gate/gate -gid 1
    ```

5. Build and run test_game:
    ```bash
    make test_game
    examples/test_game/test_game -gid 1
    ```

6. Build and run test_client:
    ```bash
    make test_client
    examples/test_client/test_client -N 500
    ```


