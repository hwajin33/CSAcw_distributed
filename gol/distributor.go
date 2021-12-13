package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"

	//"os"
	"uk.ac.bris.cs/gameoflife/util"

	//"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell{

	var aliveCells []util.Cell
	//aliveCells := []util.Cell{}
	var cell util.Cell

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				cell.X = x
				cell.Y = y
				aliveCells = append(aliveCells, cell)
			}
		}
	}
	return aliveCells
}

var server = flag.String("server","184.72.86.73:8030","IP:port string to connect to as server")

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	currentWorld := make([][]byte, p.ImageHeight)
	for i := range currentWorld {
		currentWorld[i] = make([]byte, p.ImageWidth)
	}


	filename := fmt.Sprintf("%vx%v", p.ImageWidth, p.ImageHeight)

	c.ioCommand <- ioInput
	c.ioFilename <- filename

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			currentWorld[y][x] = <-c.ioInput
		}
	}

	// we want all the turns to be processed on the remote node, and we want to get the result back
	turn := p.Turns

	//server := flag.String("server","127.0.0.1:8030","IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	//client, _ := rpc.Dial("tcp","127.0.0.1:8060")
	////panic(err) <- this not include
	//defer client.Close()

	testWorld := makeCall(*client, currentWorld, turn, p.ImageHeight, p.ImageWidth)

	aliveCell := calculateAliveCells(p, testWorld)
	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{turn, aliveCell}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

// could add more parameters
func makeCall(client rpc.Client, world [][]byte, turn int, imageHeight int, imageWidth int) [][]byte {
	request := stubs.Request{World: world, NumberOfTurns: turn, HeightImage: imageHeight, WidthImage: imageWidth}
	// new() makes a pointer
	response := new(stubs.Response)
	// response needs to be a pointer to a type when request is just a type itself
	client.Call(stubs.GameOfLife, request, response)
	return response.World
}