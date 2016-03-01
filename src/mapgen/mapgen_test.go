package mapgen

import (
	"math"
	"testing"
)

func TestMapgen(t *testing.T) {
	for i := 1; i < 20; i++ {
		field := GenerateMap(i)
		for _, currCircleOne := range field.Cells {
			for _, currCircleTwo := range field.Cells {
				if currCircleOne != currCircleTwo {
					deltaX := currCircleOne.Location.X - currCircleTwo.Location.X
					deltaY := currCircleOne.Location.Y - currCircleTwo.Location.Y
					currentDistance := math.Sqrt((math.Pow(float64(deltaX), 2) + math.Pow(float64(deltaY), 2)))
					if currentDistance < CellMaximumSize*DistanceFactor {
						t.Error("Expected distance more than ", CellMaximumSize*DistanceFactor, " got ", currentDistance, " instead")
					}
				}
			}
		}
		for _, currentPlayer := range field.StartCellIdxs {
			cellSize := field.Cells[currentPlayer].Radius
			if cellSize != PlayerCellDefaultSize {
				t.Error("Expected Cellsize of ", PlayerCellDefaultSize, " got ", cellSize, " instead ")
			}
		}
	}
}

func BenchmarkMapgen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateMap(20)
	}
}
