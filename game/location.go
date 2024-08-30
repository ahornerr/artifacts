package game

import "math"

type Location struct {
	Name string
	X    int
	Y    int
}

func (l Location) DistanceTo(other Location) float64 {
	// Euclidean
	//math.Sqrt(math.Pow(float64(location.X-currentX), 2) + math.Pow(float64(location.Y-currentY), 2))

	// Manhattan
	return math.Abs(float64(l.X-other.X)) + math.Abs(float64(l.Y-other.Y))
}
