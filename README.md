# GoWorld [![GoDoc](https://godoc.org/github.com/xiaonanln/goworld?status.png)](https://godoc.org/github.com/xiaonanln/goworld) [![Build Status](https://api.travis-ci.org/xiaonanln/goworld.svg?branch=master)](https://travis-ci.org/xiaonanln/goworld) [![Go Report Card](https://goreportcard.com/badge/github.com/xiaonanln/goworld)](https://goreportcard.com/report/github.com/xiaonanln/goworld) [![ApacheLicense](https://img.shields.io/badge/license-APACHE%20License-blue.svg)](https://raw.githubusercontent.com/xiaonanln/goworld/master/LICENSE)

_**Scalable Distributed Game Server Engine with Hot Reload in Golang**_
---------------------------------------
  * [Features](#features)
  * [Architecture](#architecture)
  * [Introduction](#introduction)
  * [Get GoWorld](#get-goworld)
  * [Manage GoWorld using goworld.py](#manage-goworld-using-goworldpy)
  * [Run Example Chatroom Client](#run-example-chatroom-client)
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
```bash
$ go get github.com/xiaonanln/goworld
```

### Install dependencies

**Windows**: `install-deps-win.bat`

**Linux**: `make install-deps`

## Manage GoWorld using goworld.py

goworld.py is a script for managing goworld server easily. Running goworld.py requires python 2.7.x and psutil module.
We can use goworld.py to build, start, stop and reload game servers. 

```bash
$ pip install psutil
```

**Build Dispatcher & Gate:**

```bash
$ python goworld.py build dispatcher gate
```

**Build Example Chatroom Server:**
```bash
$ python goworld.py build examples/chatroom_demo
```

**Start Example Chatroom Server: (dispatcher -> game -> gate)**
```bash
$ python goworld.py start examples/chatroom_demo
``` 

**Stop Game Server (gate -> game -> dispatcher):**
```bash
$ python goworld.py stop
```

**Reload Game Servers:**
```bash
$ python goworld.py reload
```
Reload will reboot game processes with the current executive while presearving all game server states. 

## Run Example Chatroom Client ##

The client for chatroom demo is hosted at [github.com/xiaonanln/goworld-chatroom-demo-client](https://github.com/xiaonanln/goworld-chatroom-demo-client).
The project was created and built in [Cocos Creater 1.5](http://www.cocos2d-x.org/). A running server & client demo can also be found at [http://goworldgs.com/chatclient/](http://goworldgs.com/chatclient/).
