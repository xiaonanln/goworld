package entity

type AOICalculator interface {
	Enter(aoi *AOI, pos Position)
	Leave(aoi *AOI)
	Move(aoi *AOI, newPos Position)
	Interested(aoi *AOI) AOISet
}

type XZListAOICalculator struct {
	xSweepList *xAOIList
	zSweepList *zAOIList
}

func newSweepAndPruneAOICalculator() *XZListAOICalculator {
	return &XZListAOICalculator{
		xSweepList: newXAOIList(),
		zSweepList: newZAOIList(),
	}
}

func (cal *XZListAOICalculator) Enter(aoi *AOI, pos Position) {
	aoi.pos = pos
	cal.xSweepList.Insert(aoi)
	cal.zSweepList.Insert(aoi)

	//for otherAOI := range cal.Interested(entity) {
	//	// interest each other
	//	otherEntity := otherAOI.getEntity()
	//	entity.interest(otherEntity)
	//	otherEntity.interest(entity)
	//}
}

func (cal *XZListAOICalculator) Leave(aoi *AOI) {
	cal.xSweepList.Remove(aoi)
	cal.zSweepList.Remove(aoi)
}

func (cal *XZListAOICalculator) Move(aoi *AOI, pos Position) {
	oldPos := aoi.pos
	aoi.pos = pos
	if oldPos.X != pos.X {
		cal.xSweepList.Move(aoi, oldPos.X)
	}
	if oldPos.Z != pos.Z {
		cal.zSweepList.Move(aoi, oldPos.Z)
	}

}

func (cal *XZListAOICalculator) Interested(aoi *AOI) AOISet {
	s1 := cal.xSweepList.Interested(aoi)
	s2 := cal.zSweepList.Interested(aoi)
	interestedAOIs := s1.Join(s2)
	return interestedAOIs
}

type aoiListOperator interface {
	GetCoord(aoi *AOI) Coord
	//SetCoord(aoi *AOI) Coord
	GetNext(aoi *AOI) *AOI
	SetNext(aoi *AOI, next *AOI)
	GetPrev(aoi *AOI) *AOI
	SetPrev(aoi *AOI, prev *AOI)
}
