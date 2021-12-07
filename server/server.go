package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
)

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

// similar to the update world (something like that) function to update the world for a single iteration

type GameOfLifeOperations struct {}

// ProcessTurns MakeGameOfLife Reverse have a for loop inside that iterates over the # of iterations specified in the Request struct
// at the end return by the Response pointer
// need a method to update the world for a single iteration (similar as the ReverseString())
// need exported method to access directly from an RPC call -> only by exported methods
func (s *GameOfLifeOperations) ProcessTurns(req stubs.Request, res *stubs.Response) (err error) {
	// have a for loop that iterates over the # of iterations specified in the Request struct
	// at the end returned by the Response pointer
	for turn := 0; turn < req.NumberOfTurns; turn++ {
		req.World = calculateNextState(req.HeightImage, req.WidthImage, req.World)
	}
	res.World = req.World
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
