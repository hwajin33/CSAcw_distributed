package gol

import (
	"flag"
	"fmt"
	"net/rpc"
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

func calculateAliveCells(imageHeight int, imageWidth int, world [][]byte) []util.Cell {

	var aliveCells []util.Cell
	//aliveCells := []util.Cell{}
	var cell util.Cell

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if world[y][x] == 255 {
				cell.X = x
				cell.Y = y
				aliveCells = append(aliveCells, cell)
			}
		}
	}
	return aliveCells
}

//func totalAliveCells(p Params, currentWorld [][]byte) []util.Cell {
//	var cells = calculateAliveCells(p, currentWorld)
//}

func saveImage(p Params, c distributorChannels, world [][]byte, turn int) {
	filename := fmt.Sprintf("%vx%vx%d", p.ImageHeight, p.ImageWidth, turn)

	c.ioCommand <- ioOutput
	c.ioFilename <- filename
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- world[y][x]
		}
	}
}

var server = flag.String("server","127.0.0.1:8011","IP:port string to connect to as server")


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

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if currentWorld [y][x] == 255 {
				cell := util.Cell{X: x, Y: y}
				c.events <- CellFlipped{CompletedTurns: 0, Cell: cell}
			}
		}
	}


	// Step 2 - make an RPC call to the server, server gets that request and calculates the alive cells
	// and reports that to the local

	// we want all the turns to be processed on the remote node, and we want to get the result back
	turn := p.Turns

	//mutex := new(sync.Mutex)
	done := make(chan bool)
	//aliveCells := len(calculateAliveCells(p.ImageHeight, p.ImageWidth, currentWorld))
	//tickerTurn := 0

	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	// not executing the goroutine
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// not updating the current world
					//mutex.Lock()
					//aliveCells := len(calculateAliveCells(p.ImageHeight, p.ImageWidth, currentWorld))
					currentTurn, currentAliveCells := makeCellCountCall(client)
					//mutex.Unlock()
					// getting the current turn and the # of live cells from the server and passing down to the event channel.
					c.events <- AliveCellsCount{CompletedTurns: currentTurn, CellsCount: currentAliveCells}
					//c.events <- TurnComplete{CompletedTurns: cellCountWorld.Turn}
					//c.events <- FinalTurnComplete{CompletedTurns: turn}
				//}
				//tickerTurn++
			default:
			}
		}
	}()

	callWorld := makeCall(client, currentWorld, turn, p.ImageHeight, p.ImageWidth)
	done <- true

	saveImage(p, c, currentWorld, turn)

	//aliveCell := callCellCount(*client, callWorld, p.Turns, aliveCells, p.ImageHeight, p.ImageWidth)
		//calculateAliveCells(p, callWorld)
	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{turn, calculateAliveCells(p.ImageHeight, p.ImageWidth, callWorld)}

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

func callCellCount(client rpc.Client, world [][]byte, turn int, aliveCells int, imageHeight int, imageWidth int) *stubs.Response {
	request := stubs.Request{World: world, NumberOfTurns: turn, NumberOfAliveCells: aliveCells, HeightImage: imageHeight, WidthImage: imageWidth}
	response := new(stubs.Response)
	client.Call(stubs.CountAliveCells, request, response)
	return response
}

func makeCellCountCall(client *rpc.Client) (turn int, numberOfAliveCells int) {
	request := stubs.CellCountRequest{}
	response := new(stubs.CellCountResponse)
	client.Call(stubs.CountAliveCells, request, response)
	return response.Turn, response.AliveCells
}