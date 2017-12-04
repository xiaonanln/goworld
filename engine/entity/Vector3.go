package entity

import (
	"fmt"
	"math"
)

// Coord is the of coordinations entity position (x, y, z)
type Coord float32

// Vector3 is type of entity position
type Vector3 struct {
	X Coord
	Y Coord
	Z Coord
}

func (p Vector3) String() string {
	return fmt.Sprintf("(%.2f, %.2f, %.2f)", p.X, p.Y, p.Z)
}

// DistanceTo calculates distance between two positions
func (p Vector3) DistanceTo(o Vector3) Coord {
	dx := p.X - o.X
	dy := p.Y - o.Y
	dz := p.Z - o.Z
	return Coord(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Sub calculates Vector3 p - Vector3 o
func (p Vector3) Sub(o Vector3) Vector3 {
	return Vector3{p.X - o.X, p.Y - o.Y, p.Z - o.Z}
}

func (p Vector3) Add(o Vector3) Vector3 {
	return Vector3{p.X + o.X, p.Y + o.Y, p.Z + o.Z}
}

// Mul calculates Vector3 p * m
func (p Vector3) Mul(m Coord) Vector3 {
	return Vector3{p.X * m, p.Y * m, p.Z * m}
}

// DirToYaw convert direction represented by Vector3 to Yaw
func (dir Vector3) DirToYaw() Yaw {
	dir.Normalize()

	yaw := math.Acos(float64(dir.X))
	if dir.Z < 0 {
		yaw = math.Pi*2 - yaw
	}

	yaw = yaw / math.Pi * 180 // convert to angle

	if yaw <= 90 {
		yaw = 90 - yaw
	} else {
		yaw = 90 + (360 - yaw)
	}

	return Yaw(yaw)
}

func (p *Vector3) Normalize() {
	d := Coord(math.Sqrt(float64(p.X*p.X + p.Y + p.Y + p.Z*p.Z)))
	if d == 0 {
		return
	}
	p.X /= d
	p.Y /= d
	p.Z /= d
}

func (p Vector3) Normalized() Vector3 {
	p.Normalize()
	return p
}
