package entity

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestXAOIList_Insert(t *testing.T) {
	for i := 0; i < 10000; i++ {
		N := rand.Intn(100)
		list := newXAOIList()
		for j := 0; j < N; j++ {
			list.Insert(randAOI())
		}
		// make sure list is valid
		checkList(t, list, N)
	}
}

func TestXAOIList_Remove(t *testing.T) {
	for i := 0; i < 10000; i++ {
		N1 := rand.Intn(100)
		N2 := rand.Intn(100)
		remove := []*AOI{}
		list := newXAOIList()
		for j := 0; j < N1; j++ {
			aoi := randAOI()
			remove = append(remove, aoi)
			list.Insert(aoi)
		}

		for j := 0; j < N2; j++ {
			list.Insert(randAOI())
		}

		// make sure list is valid
		for _, aoi := range remove {
			list.Remove(aoi)
		}
		checkList(t, list, N2)
	}
}

func TestXAOIList_Move(t *testing.T) {
	for i := 0; i < 1000; i++ {
		aois := []*AOI{}
		list := newXAOIList()
		N := 1 + rand.Intn(100)
		for j := 0; j < N; j++ {
			aoi := randAOI()
			aois = append(aois, aoi)

			list.Insert(aoi)
		}

		for r := 0; r < 100; r++ {
			aoi := aois[rand.Intn(len(aois))]
			oldCoord := aoi.pos.X
			newCoord := Coord(rand.Intn(100))
			aoi.pos.X = newCoord
			list.Move(aoi, oldCoord)
			checkList(t, list, N)
		}
	}
}

func TestXAOIList_Interested(t *testing.T) {
	for i := 0; i < 1000; i++ {
		aois := []*AOI{}
		list := newXAOIList()
		N := 1 + rand.Intn(100)
		for j := 0; j < N; j++ {
			aoi := randAOI()
			aois = append(aois, aoi)

			list.Insert(aoi)
		}
		checkList(t, list, N)

		for r := 0; r < 10; r++ {
			aoi := aois[rand.Intn(len(aois))]
			list.Mark(aoi)

			for _, other := range aois {
				if other == aoi {
					continue
				}

				if other.markVal == 1 {
					if math.Abs(float64(aoi.pos.X-other.pos.X)) > _DEFAULT_AOI_DISTANCE {
						t.Fail()
					}
					other.markVal = 0
				} else {
					if math.Abs(float64(aoi.pos.X-other.pos.X)) <= _DEFAULT_AOI_DISTANCE {
						t.Fail()
					}
				}
			}
		}
	}
}

func randAOI() *AOI {
	return &AOI{
		pos: Position{
			X: Coord(rand.Intn(100)),
			Y: Coord(rand.Intn(100)),
			Z: Coord(rand.Intn(100)),
		},
	}
}

func checkList(t *testing.T, list *xAOIList, N int) {
	if list.head != nil {
		if list.head.xPrev != nil {
			t.Errorf("head's prev is not nil")
		}
	}

	if list.tail != nil {
		if list.tail.xNext != nil {
			t.Errorf("tail's next is not nil")
		}
	}

	if (list.head == nil) != (list.tail == nil) {
		t.Errorf("invalid head & tail")
	}

	p := list.head
	var last *AOI

	for i := 0; i < N; i++ {
		if p == nil {
			t.Errorf("unexcepted nil")
		}

		if last != nil {
			if last.pos.X > p.pos.X {
				t.Errorf("list is not ordered")
			}
		}

		last = p
		p = p.xNext
		if p == nil {
			if list.tail != last {
				t.Errorf("tail is wrong")
			}
		} else {
			if p.xPrev != last {
				t.Errorf("prev is wrong")
			}
		}
	}

	if p != nil {
		t.Errorf("unexpected not nil ")
	}
}
