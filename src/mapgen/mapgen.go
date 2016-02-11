package mapgen

//func main() {
// var channels []Channel  // an empty list
//channels = append(channels, Channel{name:"some channel name"})
//channels = append(channels, Channel{name:"some other name"})

import (
	"math"
	"math/rand"
)

const (
	CellMinimumSize  = 4
	CellMaximumSize  = 24
	MaximumXPosition = 800
	MaximumYPosition = 800
	NumberOfCells    = 50
	DistanceFactor   = 1
)

type Cell struct {
	xPosition int32
	yPosition int32
	size      int32
}

type Map struct {
	cells   []Cell
	fitness float64
}

var maps []Map
var currentLowestFitness Map
var currentSecondBestFitness Map

func setFitnesses() {
	for _, currMap := range maps {

		if currentLowestFitness.cells == nil {
			currentLowestFitness = currMap
		} else if currentSecondBestFitness.cells == nil && currMap.fitness > currentLowestFitness.fitness {
			currentSecondBestFitness = currMap
		} else if currentSecondBestFitness.cells == nil && currMap.fitness <= currentLowestFitness.fitness {
			currentSecondBestFitness = currentLowestFitness
			currentLowestFitness = currMap
		} else if currMap.fitness < currentLowestFitness.fitness {
			currentSecondBestFitness = currentLowestFitness
			currentLowestFitness = currMap
		} else if currMap.fitness > currentLowestFitness.fitness && currMap.fitness < currentSecondBestFitness.fitness {
			currentSecondBestFitness = currMap
		}
	}
}

func generateMap() {
	//generationSuccessful := false

	// TODO: return generated map here
}

func calculateFitnesses(cells []Cell) float64 {
	for _, currCellOne := range cells {
		for _, currCellTwo := range cells {
			if currCellOne != currCellTwo {
				deltaX := currCellOne.xPosition - currCellTwo.xPosition
				deltaY := currCellOne.yPosition - currCellTwo.yPosition
				currentDistance := math.Sqrt((math.Pow(float64(deltaX), 2) + math.Pow(float64(deltaY), 2)))
				if currentDistance < CellMaximumSize*DistanceFactor {
					return 100
				}
			}
		}
	}
	return 0
}

func generateCellList() []Cell {
	var cells []Cell
	for i := 0; i < NumberOfCells; i++ {
		cells = append(cells, Cell{rand.Int31n(MaximumXPosition), rand.Int31n(MaximumYPosition), rand.Int31n(CellMaximumSize-CellMinimumSize) + CellMinimumSize})
	}
	return cells
}
