package entity

type AOICalculator interface {
	Enter(aoi *AOI, pos Position)
	Leave(aoi *AOI)
	Move(aoi *AOI, newPos Position)
	Adjust(aoi *AOI) (enter []*AOI, leave []*AOI)
}

type XZListAOICalculator struct {
	xSweepList *xAOIList
	zSweepList *zAOIList
}

func newXZListAOICalculator() *XZListAOICalculator {
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

func (cal *XZListAOICalculator) Adjust(aoi *AOI) (enter []*AOI, leave []*AOI) {
	cal.xSweepList.Mark(aoi)
	cal.zSweepList.Mark(aoi)
	// aoi marked twice are neighbors
	for neighbor := range aoi.neighbors {
		naoi := &neighbor.aoi
		if naoi.markVal == 2 {
			// neighbors kept
			naoi.markVal = -2 // mark this as neighbor
		} else { // markVal < 2
			// was neighbor, but not any more
			leave = append(leave, naoi)
		}
	}

	// travel in X list again to find all new neighbors, whose markVal == 2
	enter = cal.xSweepList.GetClearMarkedNeighbors(aoi)
	// travel in Z list again to unmark all
	cal.zSweepList.ClearMark(aoi)

	// now all marked neighbors are cleared
	// travel in neighbors
	return
}

type aoiListOperator interface {
	GetCoord(aoi *AOI) Coord
	//SetCoord(aoi *AOI) Coord
	GetNext(aoi *AOI) *AOI
	SetNext(aoi *AOI, next *AOI)
	GetPrev(aoi *AOI) *AOI
	SetPrev(aoi *AOI, prev *AOI)
}
