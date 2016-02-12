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
	CellMinimumSize           = 4
	CellMaximumSize           = 24
	MaximumXPosition          = 800
	MaximumYPosition          = 800
	NumberOfCells             = 50
	DistanceFactor            = 1
	NumberOfMapsPerGeneration = 8
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
var currentLowestFitness *Map
var currentSecondBestFitness *Map

func setFitnesses() {
	for _, currMap := range maps {

		if currentLowestFitness == nil {
			currentLowestFitness = &currMap
		} else if currentSecondBestFitness == nil && currMap.fitness > currentLowestFitness.fitness {
			currentSecondBestFitness = &currMap
		} else if currentSecondBestFitness == nil && currMap.fitness <= currentLowestFitness.fitness {
			currentSecondBestFitness = currentLowestFitness
			currentLowestFitness = &currMap
		} else if currMap.fitness < currentLowestFitness.fitness {
			currentSecondBestFitness = currentLowestFitness
			currentLowestFitness = &currMap
		} else if currMap.fitness > currentLowestFitness.fitness && currMap.fitness < currentSecondBestFitness.fitness {
			currentSecondBestFitness = &currMap
		}
	}
}

func generateMap() *Map {
	generationSuccessful := false
	for generationSuccessful == false {
		maps = nil
		currentLowestFitness.cells = nil
		currentLowestFitness.fitness = 0
		currentSecondBestFitness.cells = nil
		currentSecondBestFitness.fitness = 0
		for i := 0; i < NumberOfMapsPerGeneration; i++ {
			currentCellList := generateCellList()
			maps = append(maps, Map{currentCellList, calculateFitnesses(currentCellList)})
		}
		setFitnesses()
		numberOfGenerations := 0
		for currentLowestFitness.fitness != 0 && numberOfGenerations < 10000 {
			crossDivider := rand.Int31n(NumberOfCells)
			var childCellList []Cell
			for i := int32(0); i < crossDivider; i++ {
				if rand.Int31n(100) >= 2 {
					childCellList = append(childCellList, currentLowestFitness.cells[i])
				} else {
					childCellList = append(childCellList, Cell{rand.Int31n(MaximumXPosition), rand.Int31n(MaximumYPosition), rand.Int31n(CellMaximumSize-CellMinimumSize) + CellMinimumSize})
				}
			}
			for i := crossDivider; i < NumberOfCells; i++ {
				if rand.Int31n(100) >= 2 {
					childCellList = append(childCellList, currentSecondBestFitness.cells[i])
				} else {
					childCellList = append(childCellList, Cell{rand.Int31n(MaximumXPosition), rand.Int31n(MaximumYPosition), rand.Int31n(CellMaximumSize-CellMinimumSize) + CellMinimumSize})
				}
			}
			for i, currentMap := range maps {
				if &currentMap == currentSecondBestFitness {
					maps = append(maps[:i-1], maps[i+1:]...)
					maps = append(maps, Map{childCellList, calculateFitnesses(childCellList)})
				}
			}
			numberOfGenerations++
			if numberOfGenerations < 10000 {
				generationSuccessful = true
			}
		}
	}
	return currentLowestFitness
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
