# GoWorld - game server engine in Go

_Working towards alpha_ 

_**中文站点：http://goworldgs.com/?p=64**_

Scalable Distributed Game Server Engine with Hot Swapping for MMORPGs Written in Go

GoWorld server adopts a entity framework, in which entities represent all players, monsters, NPCs.
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

**Download goworld:**

`govendor get github.com/xiaonanln/goworld`

If you don't have govendor already, install by executing: 
`go get -u github.com/kardianos/govendor`

**Run goworld example server:**
1. Get MongoDB or Redis Running
2. Copy goworld.ini.sample to goworld.ini, and configure accordingly `cp goworld.ini.sample goworld.ini`
3. Build and run dispatcher: `cd components/dispatcher; go build; ./dispatcher`
4. Build and run test_game: `cd examples/test_game; go build; ./test_game -sid 1`
5. Build and run gate: `cd components/gate; go build; ./gate -gid 1`
6. Build and run test_client: `cd examples/test_client; go build; ./test_client -N 500`


