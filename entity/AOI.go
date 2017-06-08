package entity

type AOI struct {
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
