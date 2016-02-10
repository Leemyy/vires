package mapgen

//func main() {
// var channels []Channel  // an empty list
//channels = append(channels, Channel{name:"some channel name"})
//channels = append(channels, Channel{name:"some other name"})

const (
	CellMinimumSize = 4
	CellMaximumSize = 24
)

type Cell struct {
	xPosition uint32
	yPosition uint32
	size      uint32
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

		if currentLowestFitness == nil {
			currentLowestFitness = currMap
		} else if currentSecondBestFitness == nil && currMap.fitness > currentLowestFitness.fitness {
			currentSecondBestFitness = currMap
		} else if currentSecondBestFitness == nil && currMap.fitness <= currentLowestFitness.fitness {
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

func generateMap() Map {
	var bool generationSuccessful = false

	// TODO: return generated map here
	return nil
}

func calculateFitnesses(cells []Cell) float64 {

	// TODO: return Mapfitness here
	return nil
}
