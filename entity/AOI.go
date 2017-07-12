package entity

import (
	"fmt"
	"math"
	"unsafe"
)

type Coord float32

type Position struct {
	X Coord
	Y Coord
	Z Coord
}

func (p Position) String() string {
	return fmt.Sprintf("(%.1f, %.1f, %.1f)", p.X, p.Y, p.Z)
}

func (p Position) DistanceTo(o Position) Coord {
	dx := p.X - o.X
	dy := p.Y - o.Y
	dz := p.Z - o.Z
	return Coord(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

type AOI struct {
	pos       Position
	neighbors EntitySet
	xNext     *AOI
	xPrev     *AOI
	zNext     *AOI
	zPrev     *AOI
}

func initAOI(aoi *AOI) {
	aoi.neighbors = EntitySet{}
}

// Get the owner entity of this AOI
// This is very tricky but also effective
func (aoi *AOI) getEntity() *Entity {
	dummyEntity := (*Entity)(unsafe.Pointer(aoi))
	offset := uintptr(unsafe.Pointer(&dummyEntity.aoi)) - uintptr(unsafe.Pointer(dummyEntity))
	return (*Entity)(unsafe.Pointer((uintptr)(unsafe.Pointer(aoi)) - offset))
}

func (aoi *AOI) interest(other *Entity) {
	aoi.neighbors.Add(other)
}

func (aoi *AOI) uninterest(other *Entity) {
	aoi.neighbors.Del(other)
}

//func (sl *xAOIList) coord(aoi *AOI) Coord {
//	if sl.xorz == 0 {
//		return aoi.pos.X
//	} else {
//		return aoi.pos.Z
//	}
//}

//func (sl *xAOIList) head(aoi *AOI) *sweepListHead {
//	if sl.xorz == 0 {
//		return &aoi.sweepListHeadX
//	} else {
//		return &aoi.sweepListHeadZ
//	}
//}
//
//func (sl *xAOIList) next(aoi *AOI) *AOI {
//	if sl.xorz == 0 {
//		return aoi.sweepListHeadX.next
//	} else {
//		return aoi.sweepListHeadZ.next
//	}
//}
//
//func (sl *xAOIList) prev(aoi *AOI) *AOI {
//	if sl.xorz == 0 {
//		return aoi.sweepListHeadX.prev
//	} else {
//		return aoi.sweepListHeadZ.prev
//	}
//}

type AOISet map[*AOI]struct{}

func (aoiset AOISet) Add(aoi *AOI) {
	aoiset[aoi] = struct{}{}
}

func (aoiset AOISet) Del(aoi *AOI) {
	delete(aoiset, aoi)
}

func (aoiset AOISet) Contains(aoi *AOI) bool {
	_, ok := aoiset[aoi]
	return ok
}

func (aoiset AOISet) Join(other AOISet) AOISet {
	join := AOISet{}
	for aoi, _ := range aoiset {
		if other.Contains(aoi) {
			join.Add(aoi)
		}
	}
	return join
}
