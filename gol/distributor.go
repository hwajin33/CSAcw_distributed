package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"sync"
	"time"
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

func saveImage(p Params, c distributorChannels, world [][]byte, turn int) {
	filename := fmt.Sprintf("%vx%vx%d", p.ImageHeight, p.ImageWidth, turn)

	c.ioCommand <- ioOutput
	c.ioFilename <- filename
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				c.ioOutput <- 255
			}else {
				c.ioOutput <- 0
			}
		}
	}
	c.events <- ImageOutputComplete{turn, filename}
}

// IO reading image from the Distributor
func readImage(p Params, c distributorChannels, world [][]byte) [][]byte {
	filename := fmt.Sprintf("%vx%v", p.ImageWidth, p.ImageHeight)

	c.ioCommand <- ioInput
	c.ioFilename <- filename

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			image := <-c.ioInput
			world[y][x] = image
			if image == 255 {
				c.events <- CellFlipped{0, util.Cell{X: x, Y: y}}
			}
		}
	}
	return world
}

var server = flag.String("server","184.72.86.73:8030","IP:port string to connect to as server")

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	currentWorld := make([][]byte, p.ImageHeight)
	for i := range currentWorld {
		currentWorld[i] = make([]byte, p.ImageWidth)
	}


	// we want all the turns to be processed on the remote node, and we want to get the result back
	turn := p.Turns

	readImage(p, c, currentWorld)

	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	done := make(chan bool)
	mutex := sync.Mutex{}

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				//mutex.Lock()
				currentTurn, currentAliveCells := makeCellCountCall(client, turn, p.ImageHeight, p.ImageWidth)
				//mutex.Unlock()
				// getting the current turn and the # of live cells from the server and passing down to the event channel.
				c.events <- AliveCellsCount{CompletedTurns: currentTurn, CellsCount: currentAliveCells}
			default:
			}
		}
	}()

	mutex.Lock()
	callWorld := makeCall(client, currentWorld, turn, p.ImageHeight, p.ImageWidth)
	mutex.Unlock()

	done <- true

	saveImage(p, c, currentWorld, turn)

	aliveCell := calculateAliveCells(p, callWorld)
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
func makeCall(client *rpc.Client, world [][]byte, turn int, imageHeight int, imageWidth int) [][]byte {
	request := stubs.Request{World: world, NumberOfTurns: turn, HeightImage: imageHeight, WidthImage: imageWidth}
	// new() makes a pointer
	response := new(stubs.Response)
	// response needs to be a pointer to a type when request is just a type itself
	client.Call(stubs.GameOfLife, request, response)
	return response.World
}

func makeCellCountCall(client *rpc.Client, currentTurn int, imageHeight int, imageWidth int) (turn int, numberOfAliveCells int) {
	request := stubs.CellCountRequest{TotalTurns: currentTurn, HeightImage: imageHeight, WidthImage: imageWidth}
	response := new(stubs.CellCountResponse)
	client.Call(stubs.CountAliveCells, request, response)
	return response.Turn, response.AliveCells
}