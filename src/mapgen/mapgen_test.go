package mapgen

import (
	"math"
	"testing"
)

func TestMapgen(t *testing.T) {
	circles, playerCells := GenerateMap(4)
	if len(circles) != NumberOfCells {
		t.Error("Expected ", NumberOfCells, " Cells, got ", len(circles), " instead")
	}
	for _, currCircleOne := range circles {
		for _, currCircleTwo := range circles {
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
	for _, currentPlayer := range playerCells {
		cellSize := circles[currentPlayer].Radius
		if cellSize != PlayerCellDefaultSize {
			t.Error("Expected Cellsize of ", PlayerCellDefaultSize, " got ", cellSize, " instead ")
		}
	}
}

func BenchmarkMapgen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateMap(4)
	}
}
