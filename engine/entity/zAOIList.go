package entity

type zAOIList struct {
	head *aoi
	tail *aoi
}

func newZAOIList() *zAOIList {
	return &zAOIList{}
}

func (sl *zAOIList) Insert(aoi *aoi) {
	insertCoord := aoi.pos.Z
	if sl.head != nil {
		p := sl.head
		for p != nil && p.pos.Z < insertCoord {
			p = p.zNext
		}
		// now, p == nil or p.coord >= insertCoord
		if p == nil { // if p == nil, insert aoi at the end of list
			tail := sl.tail
			tail.zNext = aoi
			aoi.zPrev = tail
			sl.tail = aoi
		} else { // otherwise, p >= aoi, insert aoi before p
			prev := p.zPrev
			aoi.zNext = p
			p.zPrev = aoi
			aoi.zPrev = prev

			if prev != nil {
				prev.zNext = aoi
			} else { // p is the head, so aoi should be the new head
				sl.head = aoi
			}
		}
	} else {
		sl.head = aoi
		sl.tail = aoi
	}
}

func (sl *zAOIList) Remove(aoi *aoi) {
	prev := aoi.zPrev
	next := aoi.zNext
	if prev != nil {
		prev.zNext = next
		aoi.zPrev = nil
	} else {
		sl.head = next
	}
	if next != nil {
		next.zPrev = prev
		aoi.zNext = nil
	} else {
		sl.tail = prev
	}
}

func (sl *zAOIList) Move(aoi *aoi, oldCoord Coord) {
	coord := aoi.pos.Z
	if coord > oldCoord {
		// moving to next ...
		next := aoi.zNext
		if next == nil || next.pos.Z >= coord {
			// no need to adjust in list
			return
		}
		prev := aoi.zPrev
		//fmt.Println(1, prev, next, prev == nil || prev.zNext == aoi)
		if prev != nil {
			prev.zNext = next // remove aoi from list
		} else {
			sl.head = next // aoi is the head, trim it
		}
		next.zPrev = prev

		//fmt.Println(2, prev, next, prev == nil || prev.zNext == next)
		prev, next = next, next.zNext
		for next != nil && next.pos.Z < coord {
			prev, next = next, next.zNext
			//fmt.Println(2, prev, next, prev == nil || prev.zNext == next)
		}
		//fmt.Println(3, prev, next)
		// no we have prev.X < coord && (next == nil || next.X >= coord), so insert between prev and next
		prev.zNext = aoi
		aoi.zPrev = prev
		if next != nil {
			next.zPrev = aoi
		} else {
			sl.tail = aoi
		}
		aoi.zNext = next

		//fmt.Println(4)
	} else {
		// moving to prev ...
		prev := aoi.zPrev
		if prev == nil || prev.pos.Z <= coord {
			// no need to adjust in list
			return
		}

		next := aoi.zNext
		if next != nil {
			next.zPrev = prev
		} else {
			sl.tail = prev // aoi is the head, trim it
		}
		prev.zNext = next // remove aoi from list

		next, prev = prev, prev.zPrev
		for prev != nil && prev.pos.Z > coord {
			next, prev = prev, prev.zPrev
		}
		// no we have next.X > coord && (prev == nil || prev.X <= coord), so insert between prev and next
		next.zPrev = aoi
		aoi.zNext = next
		if prev != nil {
			prev.zNext = aoi
		} else {
			sl.head = aoi
		}
		aoi.zPrev = prev
	}
}

func (sl *zAOIList) Mark(aoi *aoi) {
	prev := aoi.zPrev
	coord := aoi.pos.Z

	minCoord := coord - _DEFAULT_AOI_DISTANCE
	for prev != nil && prev.pos.Z >= minCoord {
		prev.markVal += 1
		prev = prev.zPrev
	}

	next := aoi.zNext
	maxCoord := coord + _DEFAULT_AOI_DISTANCE
	for next != nil && next.pos.Z <= maxCoord {
		next.markVal += 1
		next = next.zNext
	}
}

func (sl *zAOIList) ClearMark(aoi *aoi) {
	prev := aoi.zPrev
	coord := aoi.pos.Z

	minCoord := coord - _DEFAULT_AOI_DISTANCE
	for prev != nil && prev.pos.Z >= minCoord {
		prev.markVal = 0
		prev = prev.zPrev
	}

	next := aoi.zNext
	maxCoord := coord + _DEFAULT_AOI_DISTANCE
	for next != nil && next.pos.Z <= maxCoord {
		next.markVal = 0
		next = next.zNext
	}
}
