package game

import (

	"fmt"
)
//func main() {
// var channels []Channel  // an empty list
  //channels = append(channels, Channel{name:"some channel name"})
  //channels = append(channels, Channel{name:"some other name"})


type Cell struct{
	var xPosition uint32
	var yPosition uint32
	var size uint32 
}

type Map struct{
	var cells []Cell
	var fitness float64
}

	const Cell_minimum_size uint8 = 4
	const Cell_maximum_size uint8 = 24

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


	