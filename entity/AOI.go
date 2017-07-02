package entity

import "math"

type Coord float32

type Position struct {
	X Coord
	Y Coord
	Z Coord
}

func (p Position) DistanceTo(o Position) Coord {
	dx := p.X - o.X
	dy := p.Y - o.Y
	dz := p.Z - o.Z
	return Coord(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

type AOI struct {
	pos            Position
	neighbors      EntitySet

}

func initAOI(aoi *AOI) {
	aoi.neighbors = EntitySet{}
}

func (aoi *AOI) interest(other *Entity) {
	aoi.neighbors.Add(other)
}

func (aoi *AOI) uninterest(other *Entity) {
	aoi.neighbors.Del(other)
}

type sweepListHead struct {
	prev *AOI
	next *AOI
}

type SweepList struct {
	head *AOI
	tail *AOI
	xorz byte
}

func newSweepList(xorz byte) *SweepList {
	return &SweepList{
		head: nil, tail: nil, xorz: xorz,
	}
}

func (sl *SweepList) Add(aoi *AOI, v Coord) {
	insertCoord := aoi.pos.X
	if sl.head != nil {
		p := sl.head
		for p != nil && p.pos.X < insertCoord {
			p = p.sweepListHeadX.next
		}
		// now, p == nil or p.coord >= insertCoord
		// if p == nil, insert aoi at the end of list
		if p == nil {
			tail := sl.tail
			sl.tail.sweepListHeadX.next =
		}
	} else {
		sl.head = aoi
		sl.tail = aoi
	}
}

//func (sl *SweepList) coord(aoi *AOI) Coord {
//	if sl.xorz == 0 {
//		return aoi.pos.X
//	} else {
//		return aoi.pos.Z
//	}
//}

//func (sl *SweepList) head(aoi *AOI) *sweepListHead {
//	if sl.xorz == 0 {
//		return &aoi.sweepListHeadX
//	} else {
//		return &aoi.sweepListHeadZ
//	}
//}
//
//func (sl *SweepList) next(aoi *AOI) *AOI {
//	if sl.xorz == 0 {
//		return aoi.sweepListHeadX.next
//	} else {
//		return aoi.sweepListHeadZ.next
//	}
//}
//
//func (sl *SweepList) prev(aoi *AOI) *AOI {
//	if sl.xorz == 0 {
//		return aoi.sweepListHeadX.prev
//	} else {
//		return aoi.sweepListHeadZ.prev
//	}
//}
