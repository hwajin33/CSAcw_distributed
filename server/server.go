package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var getTurn int
var getWorld [][]byte
var mutex = new(sync.Mutex)
//var aliveCells []util.Cell

const alive = 255
const dead = 0

func getAliveNeighbour(imageHeight int, imageWidth int, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if !(i == 0 && j == 0){
				if world[modulus(y+i, imageHeight)][modulus(x+j, imageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

func modulus(i, m int) int {
	return (i + m) % m
}

func calculateNextState(imageHeight int, imageWidth int, currentTurnWorld [][]byte) [][]byte {
	newWorld := make([][]byte, imageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, imageWidth)
	}

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			neighbours := getAliveNeighbour(imageHeight, imageWidth, x, y, currentTurnWorld)
			if currentTurnWorld[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

func calculateAliveCells(imageHeight int, imageWidth int, world [][]byte) []util.Cell{
	// Prob.: had problems sending a new slice for every alive cells calculated, it was previously appending the new cell slices
	// to the cell slices it had before.
	var cellAlive []util.Cell

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if world[y][x] == 255 {
				cellAlive = append(cellAlive, util.Cell{X: x, Y: y})
			}
		}
	}
	return cellAlive
}

// need a method to update the world for a single iteration (similar as the ReverseString())
// need exported method to access directly from an RPC call -> only by exported methods

// similar to the update world (something like that) function to update the world for a single iteration

type GameOfLifeOperations struct {}


// ProcessTurns processing the GOL on the requested world from the client and returning the processed values as response pointer in the end
func (s *GameOfLifeOperations) ProcessTurns(req stubs.Request, res *stubs.Response) (err error) {
	// have a for loop that iterates over the # of iterations specified in the Request struct
	mutex.Lock()
	// let the global variables to have the value of the requested values
	getWorld = req.World
	//aliveCells = req.AliveCells
	getTurn = req.Turns
	mutex.Unlock()
	// iterates over the # of iterations specified in the Request struct
	for getTurn < req.NumberOfTurns {
		mutex.Lock()
		// calculating and updating the requested world
		getWorld = calculateNextState(req.HeightImage, req.WidthImage, getWorld)
		//aliveCells = calculateAliveCells(req.HeightImage, req.WidthImage, getWorld)
		getTurn++
		mutex.Unlock()
	}
	mutex.Lock()
	// letting the response values to have the processed values & responses are return as a pointer
	res.World = getWorld
	//res.AliveCells = aliveCells
	res.Turn = getTurn
	mutex.Unlock()
	return
}

// ReportAliveCells reporting the number of alive cells with its turns
func (s *GameOfLifeOperations) ReportAliveCells(req stubs.CellCountRequest, res *stubs.CellCountResponse) (err error) {
	mutex.Lock()
	res.Turn = getTurn
	// letting the response value to have the number of alive cells
	// height: just the length of the world
	// width: just the horizontal length of the fist vector
	res.AliveCells = len(calculateAliveCells(len(getWorld), len(getWorld[0]), getWorld))
	mutex.Unlock()
	return
}

func main(){
	pAddr := flag.String("port","8030","Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}