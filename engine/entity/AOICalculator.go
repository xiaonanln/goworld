package entity

// AOICalculator defines interface for aoi Calculators
type AOICalculator interface {
	// Let Entity aoi enter at specified position
	Enter(aoi *aoi, pos Position)
	// Let Entity aoi leave
	Leave(aoi *aoi)
	// Let Entity aoi move
	Move(aoi *aoi, newPos Position)
	// Calculate EntityAOI Adjustment of neighbors
	Adjust(aoi *aoi) (enter []*aoi, leave []*aoi)
}

// XZListAOICalculator is an implementation of AOICalculator using XZ lists
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

// Enter is called when Entity enters Space
func (cal *XZListAOICalculator) Enter(aoi *aoi, pos Position) {
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

// Leave is called when Entity leaves Space
func (cal *XZListAOICalculator) Leave(aoi *aoi) {
	cal.xSweepList.Remove(aoi)
	cal.zSweepList.Remove(aoi)
}

// Move is called when Entity moves in Space
func (cal *XZListAOICalculator) Move(aoi *aoi, pos Position) {
	oldPos := aoi.pos
	aoi.pos = pos
	if oldPos.X != pos.X {
		cal.xSweepList.Move(aoi, oldPos.X)
	}
	if oldPos.Z != pos.Z {
		cal.zSweepList.Move(aoi, oldPos.Z)
	}
}

// Adjust is called by Entity to adjust neighbors
func (cal *XZListAOICalculator) Adjust(aoi *aoi) (enter []*aoi, leave []*aoi) {
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
	GetCoord(aoi *aoi) Coord
	//SetCoord(aoi *aoi) Coord
	GetNext(aoi *aoi) *aoi
	SetNext(aoi *aoi, next *aoi)
	GetPrev(aoi *aoi) *aoi
	SetPrev(aoi *aoi, prev *aoi)
}
