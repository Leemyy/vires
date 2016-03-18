package mapgen

import (
	"math"
	"math/rand"
	"time"

	"github.com/mhuisi/vires/src/cfg"
	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/vec"
)

var (
	// radius needed to be able to place all start cells
	safeLength = 6 * cfg.Mapgen.MinStartCellDist
	// space needed to avoid overlapping with
	// another start cell
	safeSpace = safeLength * safeLength
	// radius for area considered close to cell
	nearRadius = 3*cfg.Mapgen.MaxRadius + cfg.Mapgen.Gap
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

func fieldSize(nplayers int) vec.V {
	neededSpace := safeSpace * float64(nplayers)
	sideLength := math.Sqrt(neededSpace)
	return vec.V{sideLength, sideLength}
}

func randCoord(max float64) float64 {
	return randRangeF(0, max)
}

func (f *Field) randLoc() vec.V {
	return vec.V{randCoord(f.Size.X), randCoord(f.Size.Y)}
}

func (f *Field) generateStartCells(nplayers int) {
	cells := f.Cells
	startCells := f.StartCells
	startCellRadius := randRangeF(cfg.Mapgen.MinStartCellRadius, cfg.Mapgen.MaxRadius)
	for len(cells) < nplayers {
		c := &ent.Circle{
			Location: f.randLoc(),
			Radius:   startCellRadius,
		}
		tooClose := false
		for c2 := range f.StartCells {
			if vec.Dist(c2.Location, c.Location) < cfg.Mapgen.MinStartCellDist {
				tooClose = true
				break
			}
		}
		if !tooClose {
			cells[c] = struct{}{}
			startCells[c] = struct{}{}
		}
	}
}

func (f *Field) overlaps(cell *ent.Circle) bool {
	for c := range f.Cells {
		if vec.Dist(c.Location, cell.Location) < c.Radius+cell.Radius+cfg.Mapgen.Gap {
			return true
		}
	}
	return false
}

func (f *Field) generateNeutralCells(nplayers, cellsPerPlayer int) {
	cells := f.Cells
	neutralCells := nplayers * (cellsPerPlayer - 1)
	for i := 0; i < neutralCells; i++ {
		c := &ent.Circle{
			Location: f.randLoc(),
			Radius:   randRangeF(cfg.Mapgen.MinRadius, cfg.Mapgen.MaxRadius),
		}
		if !f.overlaps(c) {
			cells[c] = struct{}{}
		}
	}
}

func randPointInCircle(c ent.Circle, offset float64) vec.V {
	angle := rand.Float64() * 2 * math.Pi
	r := offset + (c.Radius-offset)*math.Sqrt(rand.Float64())
	l := c.Location
	x := l.X + r*math.Cos(angle)
	y := l.Y + r*math.Sin(angle)
	return vec.V{x, y}
}

func (f *Field) improveFairness() {
	cells := f.Cells
	for sc := range f.StartCells {
		n := 0
		// find the amount of cells that overlap
		// with the nearRadius circle
		for c := range cells {
			if sc != c && vec.Dist(sc.Location, c.Location) < nearRadius+c.Radius {
				n++
			}
		}
		closeCircle := ent.Circle{
			Location: sc.Location,
			Radius:   nearRadius,
		}
		// generate cells until we have enough cells
		for n < cfg.Mapgen.CellsNearStartCell {
			r := randRangeF(cfg.Mapgen.MinRadius, cfg.Mapgen.MaxRadius)
			c := &ent.Circle{
				Location: randPointInCircle(closeCircle, sc.Radius+r+cfg.Mapgen.Gap),
				Radius:   r,
			}
			// generating circles outside of the field within
			// the close circle is possible because we
			// adapt the size of the field later
			if !f.overlaps(c) {
				cells[c] = struct{}{}
				n++
			}
		}
	}
}

func (f *Field) adaptSize() {
	minx := math.MaxFloat64
	maxx := -math.MaxFloat64
	miny := math.MaxFloat64
	maxy := -math.MaxFloat64
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
	size := fieldSize(nplayers)
	f := Field{map[*ent.Circle]struct{}{}, map[*ent.Circle]struct{}{}, size}
	f.generateStartCells(nplayers)
	cellsPerPlayer := randRangeI(cfg.Mapgen.MinCellsPerPlayer, cfg.Mapgen.MaxCellsPerPlayer)
	f.generateNeutralCells(nplayers, cellsPerPlayer)
	f.improveFairness()
	f.adaptSize()
	return f
}
