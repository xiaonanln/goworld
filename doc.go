/*
GoWorld is a distributed game server engine. GoWorld adopts a Space-Entity framework for game server programming.
Entities can migrate between spaces by calling `EnterSpace`. Entities can call each other using EntityID which is a
global unique identifier for each entity. Entites can be used to represent game objects like players, monsters, NPCs, etc.

Multiprocessing

GoWorld server contains multiple processes. There should be at least 3 processes: 1 dispatcher + 1 gate + 1 game.
The gate process is responsable for handling game client connections. Currently, gate supports multiple
transmission protocol including TCP, KCP or WebSocket. It also support data compression and encryption.
The game process is where game logic actually runs. A Space will always reside in one game process where it is created.
Entities can migrate between multiple game processes by entering spaces on other game processes.
GoWorld server can scale arbitrarily by running more process.

Package goworld

goworld package is dedicated to provide GoWorld game engine APIs for developers. Most of time developers should use
functions exported by goworld package to manipulate spaces and entities. Developers can also use public methods of
Space and Entity.

Run game

GoWorld does not provide a game executable. Developers have to build their own game program. A common game program looks
like bellow:

	import "goworld"

	func main() {
		goworld.RegisterSpace(&MySpace{}) // Register a custom Space type
		// Register service entity types
		goworld.RegisterService("OnlineService", &OnlineService{})
		goworld.RegisterService("SpaceService", &SpaceService{})
		// Register Account entity type
		goworld.RegisterEntity("Account", &Account{})
		// Register Monster entity type
		goworld.RegisterEntity("Monster", &Monster{})
		// Register Player entity type
		goworld.RegisterEntity("Player", &Player{})
		// Run the game server
		goworld.Run()
	}

Basically, you need to register space type, service types and entity types and then start the endless loop of game logic.

Creating Spaces

Use goworld.CreateSpace* functions to create spaces.

Creating Entities

use goworld.CreateEntity* functions to create entities.

Loading Entities

use goworld.LoadEntity* functions to load entities from database.

Entity RPC

use goworld.Call* functions to do RPC among entities

Entity Attributes



Configuration

GoWorld uses `goworld.ini` as the default config file.

*/
package goworld
