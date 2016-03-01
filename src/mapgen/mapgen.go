package mapgen

import (
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
	NeededFitness             = 500
)

//Represents a gamefield which consists of its size,
//cells and manned cells
type Field struct {
	Size          vec.V
	Cells         []ent.Circle
	StartCellIdxs []int
}

//Represents a single cell which consists of its size,
//and position on the field
type Cell struct {
	xPosition int
	yPosition int
	size      int
}

//Represents a single map which consists of its cells and its fitness
type Map struct {
	cells   []Cell
	fitness float64
}

//Represents a single generation which consists of maps and the maps
//with the best and 2nd best fitness
type Generation struct {
	maps                     []Map
	currentLowestFitness     *Map
	currentSecondBestFitness *Map
}

//Generation struct constructor
//returns a pointer to a generation struct
func newGeneration(maps []Map, lowestFitness *Map, secondLowestFitness *Map) *Generation {
	return &Generation{
		maps:                     maps,
		currentLowestFitness:     lowestFitness,
		currentSecondBestFitness: secondLowestFitness,
	}
}

//sets the fitnesses of all maps inherited in a single generation.
func setFitnesses(generation *Generation) {
	for _, currMap := range generation.maps {

		if generation.currentLowestFitness == nil {
			generation.currentLowestFitness = &currMap
		} else if generation.currentSecondBestFitness == nil && currMap.fitness < generation.currentLowestFitness.fitness {
			generation.currentSecondBestFitness = &currMap
		} else if generation.currentSecondBestFitness == nil && currMap.fitness >= generation.currentLowestFitness.fitness {
			generation.currentSecondBestFitness = generation.currentLowestFitness
			generation.currentLowestFitness = &currMap
		} else if currMap.fitness > generation.currentLowestFitness.fitness {
			generation.currentSecondBestFitness = generation.currentLowestFitness
			generation.currentLowestFitness = &currMap
		} else if currMap.fitness < generation.currentLowestFitness.fitness && currMap.fitness > generation.currentSecondBestFitness.fitness {
			generation.currentSecondBestFitness = &currMap
		}
	}
}

//Generates a map by using a genetic algorithm.
//Returns a fieldstruct with several values concerning the map.
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
			maps[i] = Map{currentCellList, calculateFitnesses(currentCellList, vec.V{float64(maximumXPosition) / 2, float64(maximumYPosition) / 2})}
		}
		generation = newGeneration(maps, nil, nil)
		setFitnesses(generation)
		numberOfGenerations := 0
		for generation.currentLowestFitness.fitness <= NeededFitness && numberOfGenerations < 10000 {
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
					generation.maps[i] = Map{childCellList, calculateFitnesses(childCellList, vec.V{float64(maximumXPosition) / 2, float64(maximumYPosition) / 2})}
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
	return Field{vec.V{float64(maximumXPosition), float64(maximumYPosition)}, circles, playerIndex}
}

//Calculates fitness by using distancealgorithms.
//Returns fitness between 0-1000 as a float64 value.
func calculateFitnesses(cells []Cell, mapMid vec.V) float64 {
	allSmallestDistances := make([]float64, len(cells))
	var smallestDistanceToMapMid float64
	for i, currCellOne := range cells {
		var smallestDistance float64
		currentDistance := calculateDistance(currCellOne.xPosition, int(mapMid.X), currCellOne.yPosition, int(mapMid.Y))
		if smallestDistanceToMapMid == 0 || smallestDistanceToMapMid > currentDistance {
			smallestDistanceToMapMid = currentDistance
		}
		for _, currCellTwo := range cells {
			if currCellOne != currCellTwo {
				currentDistance := calculateDistance(currCellOne.xPosition, currCellTwo.xPosition, currCellOne.yPosition, currCellTwo.yPosition)
				if currentDistance <= CellMaximumSize*DistanceFactor {
					return 0
				} else if smallestDistance == 0 || smallestDistance < currentDistance {
					smallestDistance = currentDistance
				}
			}
		}
		allSmallestDistances[i] = smallestDistance
	}
	smallestDistanceToMapMid = (smallestDistanceToMapMid / (mapMid.X / 2)) * 100
	return ((getLowestValue(allSmallestDistances) / getHighestValue(allSmallestDistances)) * 1000) - smallestDistanceToMapMid
}

//Calculates the distance between two 2dimensional points.
//Returns the distance as a float64 value
func calculateDistance(xPositionOne int, xPositionTwo int, yPositionOne int, yPositionTwo int) float64 {
	deltaX := xPositionOne - xPositionTwo
	deltaY := yPositionOne - yPositionTwo
	return math.Sqrt((math.Pow(float64(deltaX), 2) + math.Pow(float64(deltaY), 2)))
}

//Returns the lowest value inside an array
func getLowestValue(values []float64) float64 {
	var lowest float64
	for _, value := range values {
		if lowest == 0 || value < lowest {
			lowest = value
		}
	}
	return lowest
}

//Returns the highest value inside an array
func getHighestValue(values []float64) float64 {
	var highest float64
	for _, value := range values {
		if highest == 0 || value > highest {
			highest = value
		}
	}
	return highest
}

//Generates a list containing a given number of cells.
//Returns a cell list
func generateCellList(maximumXPosition int, maximumYPosition, numberOfCells int) []Cell {
	var cells []Cell
	for i := 0; i < numberOfCells; i++ {
		cells = append(cells, Cell{rand.Intn(maximumXPosition), rand.Intn(maximumYPosition), (rand.Intn(CellMaximumSize-CellMinimumSize) + CellMinimumSize) / 2})
	}
	return cells
}

//Determines whether an array contains a given value
func contains(array []int, value int) bool {
	for _, a := range array {
		if a == value {
			return true
		}
	}
	return false
}
