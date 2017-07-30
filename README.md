# GoWorld - game server engine in Golang

_**Scalable Distributed Game Server Engine with Hot Swapping for MMORPGs Written in Golang**_

>  
> 
> _Working towards alpha_
>
> **QQ群：`662182346`** _**中文站点：http://goworldgs.com/?p=64**_
---------------------------------------
  * [Features](#features)
  * [Architecture](#architecture)
  * [Introduction](#introduction)
  * [Get GoWorld](#get-goworld)
  * [Run Example Server & Client](#run-example-server--client)
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
go get -u github.com/garyburd/redigo
go get -u github.com/google/btree
go get -u github.com/pkg/errors
go get -u gopkg.in/eapache/queue.v1
go get -u gopkg.in/ini.v1
go get -u gopkg.in/mgo.v2
go get -u gopkg.in/vmihailenco/msgpack.v2
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
    examples/test_game/test_game -sid 1
    ```

6. Build and run test_client:
    ```bash
    make test_client
    examples/test_client/test_client -N 500
    ```


