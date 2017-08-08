package entity

import (
	"fmt"
	"math"
	"unsafe"
)

// Coord is the of coordinations entity position (x, y, z)
type Coord float32

// Position is type of entity position
type Position struct {
	X Coord
	Y Coord
	Z Coord
}

func (p Position) String() string {
	return fmt.Sprintf("(%.1f, %.1f, %.1f)", p.X, p.Y, p.Z)
}

// DistanceTo calculates distance between two positions
func (p Position) DistanceTo(o Position) Coord {
	dx := p.X - o.X
	dy := p.Y - o.Y
	dz := p.Z - o.Z
	return Coord(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

type aoi struct {
	pos       Position
	neighbors EntitySet
	xNext     *aoi
	xPrev     *aoi
	zNext     *aoi
	zPrev     *aoi
	markVal   int
}

func initAOI(aoi *aoi) {
	aoi.neighbors = EntitySet{}
}

// Get the owner entity of this aoi
// This is very tricky but also effective
var aoiFieldOffset uintptr

func init() {
	dummyEntity := (*Entity)(unsafe.Pointer(&aoiFieldOffset))
	aoiFieldOffset = uintptr(unsafe.Pointer(&dummyEntity.aoi)) - uintptr(unsafe.Pointer(dummyEntity))
}
func (aoi *aoi) getEntity() *Entity {
	return (*Entity)(unsafe.Pointer((uintptr)(unsafe.Pointer(aoi)) - aoiFieldOffset))
}

func (aoi *aoi) interest(other *Entity) {
	aoi.neighbors.Add(other)
}

func (aoi *aoi) uninterest(other *Entity) {
	aoi.neighbors.Del(other)
}

type aoiSet map[*aoi]struct{}

func (aoiset aoiSet) Add(aoi *aoi) {
	aoiset[aoi] = struct{}{}
}

func (aoiset aoiSet) Del(aoi *aoi) {
	delete(aoiset, aoi)
}

func (aoiset aoiSet) Contains(aoi *aoi) bool {
	_, ok := aoiset[aoi]
	return ok
}

func (aoiset aoiSet) Join(other aoiSet) aoiSet {
	join := aoiSet{}
	for aoi := range aoiset {
		if other.Contains(aoi) {
			join.Add(aoi)
		}
	}
	return join
}
