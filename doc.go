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

You must register a Space type which must be "inherit" goworld.Space  using `RegisterSpace`. `RegisterSpace` must
be called exactly once, because GoWorld does not support multiple Space types.

	type MySpace struct {
		goworld.Space // Space type should always inherit from goworld.Space
		...
	}



Configuration

GoWorld uses `goworld.ini` as the default config file.

*/
package goworld
