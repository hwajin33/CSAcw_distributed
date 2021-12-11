package stubs

var GameOfLife = "GameOfLifeOperations.ProcessTurns"
var CountAliveCells = "GameOfLifeOperations.ReportAliveCells"

// Response going to have a 2D slice returning the final board state back to the local controller
type Response struct {
	World              [][]uint8
	AliveCells         int
	Turn               int
	NumberOfAliveCells int
}

type TurnResponse struct {
	Turn int
}

type CellCountResponse struct {
	AliveCells int
	Turn int
}

// Request Receive contains a 2D slice with
//the initial state of the board, # of turns, size of image to iterate the board correctly
type Request struct {
	World [][]uint8
	NumberOfTurns int
	NumberOfAliveCells int
	HeightImage int
	WidthImage int
}

type CellCountRequest struct {
	HeightImage int
	WidthImage int
}

type EmptyRequest struct {

}