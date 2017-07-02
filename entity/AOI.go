package entity

import "math"

type Float float32

type Position struct {
	X Float
	Y Float
	Z Float
}

func (p Position) DistanceTo(o Position) Float {
	dx := p.X - o.X
	dy := p.Y - o.Y
	dz := p.Z - o.Z
	return Float(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

type AOI struct {
	pos       Position
	neighbors EntitySet
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

type SweepList struct {
	head *AOI
	tail *AOI
}

func (sl *SweepList) Add(aoi *AOI, v Float) {
	if sl.head != nil {

	} else {
		sl.head = aoi
		sl.tail = aoi
	}
}
