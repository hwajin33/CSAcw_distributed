package stubs

var GameOfLife = "GameOfLifeOperations.ProcessTurns"
var CountAliveCells = "CountCellOperation.CountCells"

// Response going to have a 2D slice returning the final board state back to the local controller
type Response struct {
	World [][]byte
	NumberOfAliveCells int
	Turn int
}

// Request Receive contains a 2D slice with
//the initial state of the board, # of turns, size of image to iterate the board correctly
type Request struct {
	World [][]byte
	NumberOfTurns int
	HeightImage int
	WidthImage int
}