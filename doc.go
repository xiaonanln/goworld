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

Use goworld.CreateEntity* functions to create entities.

Loading Entities

Use goworld.LoadEntity* functions to load entities from database.

Entity RPC

Use goworld.Call* functions to do RPC among entities

Entity storage and attributes

Each entity type should override function DescribeEntityType to declare its expected behavior and all attributes,
just like bellow.

	func (a *Avatar) DescribeEntityType(desc *entity.EntityTypeDesc) {
		desc.SetPersistent(true).SetUseAOI(true, 100)
		desc.DefineAttr("name", "AllClients", "Persistent")
		desc.DefineAttr("exp", "Client", "Persistent")
		desc.DefineAttr("lastMailID", "Persistent")
		desc.DefineAttr("testListField", "AllClients")
		desc.DefineAttr("enteringNilSpace")
	}

Function SetPersistent can be used to make entities persistent. Persistent entities' attributes will be marshalled and
saved on Entity Storage (e.g. MongoDB) every configurable minutes.

Entities use attributes to store related data. Attributes can be synchronized to clients automatically.
An entity's "AllClient" attributes will be synchronized to all clients of entities where this entity is
in its AOI range. "Client" attributes wil be synchronized to own clients of entities. "Persistent" attributes will be
saved on entity storage when entities are saved periodically.
When entity is migrated from one game process to another, all attributes are marshalled and sent to the target game where
the entity will be reconstructed using attribute data.

Configuration

GoWorld uses `goworld.ini` as the default config file. Use '-configfile <path>' to use specified config file for processes.

*/
package goworld
