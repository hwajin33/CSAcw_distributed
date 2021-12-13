package stubs

import "uk.ac.bris.cs/gameoflife/util"

var GameOfLife = "GameOfLifeOperations.ProcessTurns"
var CountAliveCells = "GameOfLifeOperations.ReportAliveCells"

// Response going to have a 2D slice returning the final board state back to the local controller
type Response struct {
	World              [][]uint8
	AliveCells         []util.Cell
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
	Turns int
	NumberOfAliveCells int
	AliveCells []util.Cell
	HeightImage int
	WidthImage int
}

type CellCountRequest struct {
	TotalTurns int
	HeightImage int
	WidthImage int
}

type EmptyRequest struct {

}