package entity

type xAOIList struct {
	head *aoi
	tail *aoi
}

func newXAOIList() *xAOIList {
	return &xAOIList{}
}

func (sl *xAOIList) Insert(aoi *aoi) {
	insertCoord := aoi.pos.X
	if sl.head != nil {
		p := sl.head
		for p != nil && p.pos.X < insertCoord {
			p = p.xNext
		}
		// now, p == nil or p.coord >= insertCoord
		if p == nil { // if p == nil, insert aoi at the end of list
			tail := sl.tail
			tail.xNext = aoi
			aoi.xPrev = tail
			sl.tail = aoi
		} else { // otherwise, p >= aoi, insert aoi before p
			prev := p.xPrev
			aoi.xNext = p
			p.xPrev = aoi
			aoi.xPrev = prev

			if prev != nil {
				prev.xNext = aoi
			} else { // p is the head, so aoi should be the new head
				sl.head = aoi
			}
		}
	} else {
		sl.head = aoi
		sl.tail = aoi
	}
}

func (sl *xAOIList) Remove(aoi *aoi) {
	prev := aoi.xPrev
	next := aoi.xNext
	if prev != nil {
		prev.xNext = next
		aoi.xPrev = nil
	} else {
		sl.head = next
	}
	if next != nil {
		next.xPrev = prev
		aoi.xNext = nil
	} else {
		sl.tail = prev
	}
}

func (sl *xAOIList) Move(aoi *aoi, oldCoord Coord) {
	coord := aoi.pos.X
	if coord > oldCoord {
		// moving to next ...
		next := aoi.xNext
		if next == nil || next.pos.X >= coord {
			// no need to adjust in list
			return
		}
		prev := aoi.xPrev
		//fmt.Println(1, prev, next, prev == nil || prev.xNext == aoi)
		if prev != nil {
			prev.xNext = next // remove aoi from list
		} else {
			sl.head = next // aoi is the head, trim it
		}
		next.xPrev = prev

		//fmt.Println(2, prev, next, prev == nil || prev.xNext == next)
		prev, next = next, next.xNext
		for next != nil && next.pos.X < coord {
			prev, next = next, next.xNext
			//fmt.Println(2, prev, next, prev == nil || prev.xNext == next)
		}
		//fmt.Println(3, prev, next)
		// no we have prev.X < coord && (next == nil || next.X >= coord), so insert between prev and next
		prev.xNext = aoi
		aoi.xPrev = prev
		if next != nil {
			next.xPrev = aoi
		} else {
			sl.tail = aoi
		}
		aoi.xNext = next

		//fmt.Println(4)
	} else {
		// moving to prev ...
		prev := aoi.xPrev
		if prev == nil || prev.pos.X <= coord {
			// no need to adjust in list
			return
		}

		next := aoi.xNext
		if next != nil {
			next.xPrev = prev
		} else {
			sl.tail = prev // aoi is the head, trim it
		}
		prev.xNext = next // remove aoi from list

		next, prev = prev, prev.xPrev
		for prev != nil && prev.pos.X > coord {
			next, prev = prev, prev.xPrev
		}
		// no we have next.X > coord && (prev == nil || prev.X <= coord), so insert between prev and next
		next.xPrev = aoi
		aoi.xNext = next
		if prev != nil {
			prev.xNext = aoi
		} else {
			sl.head = aoi
		}
		aoi.xPrev = prev
	}
}

func (sl *xAOIList) Mark(aoi *aoi) {
	prev := aoi.xPrev
	coord := aoi.pos.X

	minCoord := coord - _DEFAULT_AOI_DISTANCE
	for prev != nil && prev.pos.X >= minCoord {
		prev.markVal += 1
		prev = prev.xPrev
	}

	next := aoi.xNext
	maxCoord := coord + _DEFAULT_AOI_DISTANCE
	for next != nil && next.pos.X <= maxCoord {
		next.markVal += 1
		next = next.xNext
	}
}

func (sl *xAOIList) GetClearMarkedNeighbors(aoi *aoi) (enter []*aoi) {
	prev := aoi.xPrev
	coord := aoi.pos.X
	minCoord := coord - _DEFAULT_AOI_DISTANCE
	for prev != nil && prev.pos.X >= minCoord {
		if prev.markVal == 2 {
			enter = append(enter, prev)
		}
		prev.markVal = 0
		prev = prev.xPrev
	}

	next := aoi.xNext
	maxCoord := coord + _DEFAULT_AOI_DISTANCE
	for next != nil && next.pos.X <= maxCoord {
		if next.markVal == 2 {
			enter = append(enter, next)
		}
		next.markVal = 0
		next = next.xNext
	}
	return
}
