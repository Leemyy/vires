package mapgen

import (
	"math"
	"math/rand"
	"time"

	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/vec"
)

const (
	gap                = 10
	minRadius          = 40
	maxRadius          = 120
	minStartCellRadius = 0.8 * maxRadius
	// radius needed to avoid overlapping with
	// another cell and its gap
	safeRadius = 3*maxRadius + gap
	// space needed to avoid overlapping with
	// another cell and its gap
	safeSpace = math.Pi * safeRadius * safeRadius
	// no hard limits, just rough estimates
	minCellsPerPlayer = 10
	maxCellsPerPlayer = 15
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func randRangeI(min, max int) int {
	return min + rand.Intn(max-min)
}

func randRangeF(min, max float64) float64 {
	return min + (max-min)*rand.Float64()
}

type Field struct {
	Cells      map[*ent.Circle]struct{}
	StartCells map[*ent.Circle]struct{}
	Size       vec.V
}

func fieldSize(cellsPerPlayer int) vec.V {
	neededSpace := safeSpace * float64(cellsPerPlayer)
	sideLength := math.Sqrt(neededSpace)
	return vec.V{sideLength, sideLength}
}

func randCoord(max float64) float64 {
	return randRangeF(0, max)
}

func (f *Field) randLoc() vec.V {
	return vec.V{randCoord(f.Size.X), randCoord(f.Size.Y)}
}

func tooClose(c1, c2 *ent.Circle) bool {
	return vec.Dist(c1.Location, c2.Location) < c1.Radius+c2.Radius+gap
}

func (f *Field) overlaps(cell *ent.Circle) bool {
	for c := range f.Cells {
		if tooClose(c, cell) {
			return true
		}
	}
	return false
}

func (f *Field) generateStartCells(nplayers int) {
	cells := f.Cells
	startCells := f.StartCells
	startCellRadius := randRangeF(minStartCellRadius, maxRadius)
	for len(cells) < nplayers {
		c := &ent.Circle{
			Location: f.randLoc(),
			Radius:   startCellRadius,
		}
		if !f.overlaps(c) {
			cells[c] = struct{}{}
			startCells[c] = struct{}{}
		}
	}
}

func (f *Field) generateNeutralCells(nplayers, cellsPerPlayer int) {
	cells := f.Cells
	neutralCells := nplayers * (cellsPerPlayer - 1)
	for i := 0; i < neutralCells; i++ {
		c := &ent.Circle{
			Location: f.randLoc(),
			Radius:   randRangeF(minRadius, maxRadius),
		}
		if !f.overlaps(c) {
			cells[c] = struct{}{}
		}
	}
}

func (f *Field) adaptSize() {
	minx := f.Size.X
	maxx := 0.0
	miny := f.Size.Y
	maxy := 0.0
	for c := range f.Cells {
		l := c.Location
		x := l.X
		y := l.Y
		minx = math.Min(x, minx)
		maxx = math.Max(x, maxx)
		miny = math.Min(y, miny)
		maxy = math.Max(y, maxy)
	}
	f.Size.X = maxx - minx
	f.Size.Y = maxy - miny
	for c := range f.Cells {
		c.Location.X -= minx
		c.Location.Y -= miny
	}
}

func Generate(nplayers int) Field {
	cellsPerPlayer := randRangeI(minCellsPerPlayer, maxCellsPerPlayer)
	size := fieldSize(cellsPerPlayer)
	f := Field{map[*ent.Circle]struct{}{}, map[*ent.Circle]struct{}{}, size}
	f.generateStartCells(nplayers)
	f.generateNeutralCells(nplayers, cellsPerPlayer)
	f.adaptSize()
	return f
}
