package mapgen

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/mhuisi/vires/src/ent"
	"github.com/mhuisi/vires/src/vec"
)

const (
	CellMinimumSize           = 80
	PlayerCellDefaultSize     = 140
	CellMaximumSize           = 200
	DistanceFactor            = 1.1
	NumberOfMapsPerGeneration = 8
)

type Field struct {
	Size          vec.V
	Cells         []ent.Circle
	StartCellIdxs []int
}

type Cell struct {
	xPosition int
	yPosition int
	size      int
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

func GenerateMap(numberOfPlayers int) Field {
	var maximumXPosition int
	var maximumYPosition int
	var numberOfCells int
	if numberOfPlayers > 0 {
		maximumXPosition = int(2000 + 700*math.Pow(math.Log2(float64(numberOfPlayers)), 1.5))
		maximumYPosition = int(2000 + 700*math.Pow(math.Log2(float64(numberOfPlayers)), 1.5))
		numberOfCells = int(10 + 2*math.Pow(math.Log2(float64(numberOfPlayers)), 2))
	} else {
		maximumXPosition = 8000
		maximumYPosition = 8000
		numberOfCells = 10

	}
	generationSuccessful := false
	generation := newGeneration(nil, nil, nil)
	for generationSuccessful == false {
		maps := make([]Map, NumberOfMapsPerGeneration)
		for i := 0; i < NumberOfMapsPerGeneration; i++ {
			currentCellList := generateCellList(maximumXPosition, maximumYPosition, numberOfCells)
			maps[i] = Map{currentCellList, calculateFitnesses(currentCellList)}
		}
		generation = newGeneration(maps, nil, nil)
		setFitnesses(generation)
		numberOfGenerations := 0
		for generation.currentLowestFitness.fitness != 0 && numberOfGenerations < 10000 {
			crossDivider := rand.Intn(numberOfCells)
			childCellList := make([]Cell, numberOfCells)
			for i := 0; i < crossDivider; i++ {
				if rand.Intn(100) >= 2 {
					childCellList[i] = generation.currentLowestFitness.cells[i]
				} else {
					childCellList[i] = Cell{rand.Intn(maximumXPosition), rand.Intn(maximumYPosition), (rand.Intn((CellMaximumSize - CellMinimumSize)) + CellMinimumSize) / 2}
				}
			}
			for i := crossDivider; i < numberOfCells; i++ {
				if rand.Intn(100) >= 2 {
					childCellList[i] = generation.currentSecondBestFitness.cells[i]
				} else {
					childCellList[i] = Cell{rand.Intn(maximumXPosition), rand.Intn(maximumYPosition), (rand.Intn(CellMaximumSize-CellMinimumSize) + CellMinimumSize) / 2}
				}
			}
			for i, currentMap := range generation.maps {
				if &currentMap == generation.currentSecondBestFitness {
					generation.maps[i] = Map{childCellList, calculateFitnesses(childCellList)}
				}
			}
			numberOfGenerations++
		}
		if numberOfGenerations < 10000 {
			generationSuccessful = true
		}
	}
	circles := make([]ent.Circle, numberOfCells)
	for i, currentCell := range generation.currentLowestFitness.cells {
		circles[i] = ent.Circle{vec.V{float64(currentCell.xPosition), float64(currentCell.yPosition)}, float64(currentCell.size)}
	}
	playerIndex := make([]int, numberOfPlayers)
	for i := 0; i < numberOfPlayers; i++ {
		randomNumber := rand.Intn(len(circles) - 1)
		for contains(playerIndex, randomNumber) {
			randomNumber = rand.Intn(len(circles) - 1)
		}
		circles[randomNumber].Radius = PlayerCellDefaultSize / 2
		playerIndex[i] = randomNumber
	}
	fmt.Println("Number of Players: ", numberOfPlayers, "Map X Size: ", maximumXPosition, " Map Y Size: ", maximumYPosition, " Number of Cells: ", numberOfCells)

	return Field{vec.V{float64(maximumXPosition), float64(maximumYPosition)}, circles, playerIndex}
}

func calculateFitnesses(cells []Cell) float64 {
	for _, currCellOne := range cells {
		for _, currCellTwo := range cells {
			if currCellOne != currCellTwo {
				deltaX := currCellOne.xPosition - currCellTwo.xPosition
				deltaY := currCellOne.yPosition - currCellTwo.yPosition
				currentDistance := math.Sqrt((math.Pow(float64(deltaX), 2) + math.Pow(float64(deltaY), 2)))
				if currentDistance <= CellMaximumSize*DistanceFactor {
					return 100
				}
			}
		}
	}
	return 0
}

func generateCellList(maximumXPosition int, maximumYPosition, numberOfCells int) []Cell {
	var cells []Cell
	for i := 0; i < numberOfCells; i++ {
		cells = append(cells, Cell{rand.Intn(maximumXPosition), rand.Intn(maximumYPosition), (rand.Intn(CellMaximumSize-CellMinimumSize) + CellMinimumSize) / 2})
	}
	return cells
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
