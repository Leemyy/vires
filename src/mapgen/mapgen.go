package mapgen

//func main() {
// var channels []Channel  // an empty list
//channels = append(channels, Channel{name:"some channel name"})
//channels = append(channels, Channel{name:"some other name"})

import (
	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/vec"
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

type Generation struct {
	maps                     []Map
	currentLowestFitness     *Map
	currentSecondBestFitness *Map
}

func newGeneration(maps []Map, lowestFitness *Map, secondLowestFitness *Map) *Generation {
	return &Generation{
		maps:                     maps,
		currentLowestFitness:     lowestFitness,
		currentSecondBestFitness: secondLowestFitness,
	}
}

func setFitnesses(generation *Generation) {
	for _, currMap := range generation.maps {

		if generation.currentLowestFitness == nil {
			generation.currentLowestFitness = &currMap
		} else if generation.currentSecondBestFitness == nil && currMap.fitness > generation.currentLowestFitness.fitness {
			generation.currentSecondBestFitness = &currMap
		} else if generation.currentSecondBestFitness == nil && currMap.fitness <= generation.currentLowestFitness.fitness {
			generation.currentSecondBestFitness = generation.currentLowestFitness
			generation.currentLowestFitness = &currMap
		} else if currMap.fitness < generation.currentLowestFitness.fitness {
			generation.currentSecondBestFitness = generation.currentLowestFitness
			generation.currentLowestFitness = &currMap
		} else if currMap.fitness > generation.currentLowestFitness.fitness && currMap.fitness < generation.currentSecondBestFitness.fitness {
			generation.currentSecondBestFitness = &currMap
		}
	}
}

func generateMap(numberOfPlayers int) ([]ent.Circle, []int) {
	generationSuccessful := false
	generation := newGeneration(nil, nil, nil)
	for generationSuccessful == false {
		generation := newGeneration(nil, nil, nil)
		for i := 0; i < NumberOfMapsPerGeneration; i++ {
			currentCellList := generateCellList()
			generation.maps = append(generation.maps, Map{currentCellList, calculateFitnesses(currentCellList)})

		}
		setFitnesses(generation)
		numberOfGenerations := 0
		for generation.currentLowestFitness.fitness != 0 && numberOfGenerations < 10000 {
			crossDivider := rand.Int31n(NumberOfCells)
			var childCellList []Cell
			for i := int32(0); i < crossDivider; i++ {
				if rand.Int31n(100) >= 2 {
					childCellList = append(childCellList, generation.currentLowestFitness.cells[i])
				} else {
					childCellList = append(childCellList, Cell{rand.Int31n(MaximumXPosition), rand.Int31n(MaximumYPosition), rand.Int31n(CellMaximumSize-CellMinimumSize) + CellMinimumSize})
				}
			}
			for i := crossDivider; i < NumberOfCells; i++ {
				if rand.Int31n(100) >= 2 {
					childCellList = append(childCellList, generation.currentSecondBestFitness.cells[i])
				} else {
					childCellList = append(childCellList, Cell{rand.Int31n(MaximumXPosition), rand.Int31n(MaximumYPosition), rand.Int31n(CellMaximumSize-CellMinimumSize) + CellMinimumSize})
				}
			}
			for i, currentMap := range generation.maps {
				if &currentMap == generation.currentSecondBestFitness {
					generation.maps = append(generation.maps[:i-1], generation.maps[i+1:]...)
					generation.maps = append(generation.maps, Map{childCellList, calculateFitnesses(childCellList)})
				}
			}
			numberOfGenerations++
			if numberOfGenerations < 10000 {
				generationSuccessful = true
			}
		}
	}
	var circles []ent.Circle
	for _, currentCell := range generation.currentLowestFitness.cells {
		circles = append(circles, ent.Circle{vec.V{float64(currentCell.xPosition), float64(currentCell.yPosition)}, float64(currentCell.size)})
	}
	return circles, nil
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
