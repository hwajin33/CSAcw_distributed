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

var turn int
var cell util.Cell
var aliveCells []util.Cell
var world [][]byte

var mutex = new(sync.Mutex)

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

func getAliveNeighbour(imageHeight int, imageWidth int, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if !(i == 0 && j == 0){
				if world[mod(y+i, imageHeight)][mod(x+j, imageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
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

func calculateAliveCells(imageHeight int, imageWidth int, world [][]byte) []util.Cell {
	// Prob.: had problems sending a new slice for every alive cells calculated, it was previously appending the new cell slices
	// to the cell slices it had before.
	var alive []util.Cell
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if world[y][x] == 255 {
				alive = append(alive, util.Cell{X: x, Y: y})
			}
		}
	}
	return alive
}

//func wholeAliveCells(imageHeight int, imageWidth int, world [][]byte) int {
//
//	finalAliveCells := len(&calculateAiveCells)
//	return len(finalAliveCells)
//}
//
//func totalAliveCells(p Params, currentWorld [][]byte) []util.Cell {
//	var cells = calculateAliveCells(p, currentWorld)
//}

// similar to the update world (something like that) function to update the world for a single iteration

type GameOfLifeOperations struct {}

type CountCellOperation struct {}


// ProcessTurns MakeGameOfLife Reverse have a for loop inside that iterates over the # of iterations specified in the Request struct
// at the end return by the Response pointer
// need a method to update the world for a single iteration (similar as the ReverseString())
// need exported method to access directly from an RPC call -> only by exported methods
func (s *GameOfLifeOperations) ProcessTurns(req stubs.Request, res *stubs.Response) (err error) {
	// have a for loop that iterates over the # of iterations specified in the Request struct
	// at the end returned by the Response pointer
	world = req.World
	for turn < req.NumberOfTurns {
		mutex.Lock()
		world = calculateNextState(req.HeightImage, req.WidthImage, world)
		turn++
		mutex.Unlock()
	}
	res.World = world
	return
}

func (s *GameOfLifeOperations) ReportAliveCells(req stubs.CellCountRequest, res *stubs.CellCountResponse) (err error) {
	mutex.Lock()
	res.Turn = turn
	res.AliveCells = len(calculateAliveCells(len(world), len(world[0]), world))
	mutex.Unlock()
	return
}

func main(){
	pAddr := flag.String("port","8091","Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
