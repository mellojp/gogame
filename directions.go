package main

type PlayerDir uint8

const (
	Up PlayerDir = iota
	Down
	Left
	Right
	UpLeft
	UpRight
	DownLeft
	DownRight
)

type Vec struct {
	X int
	Y int
}

var VectorToDir = map[Vec]PlayerDir{
	{0, -1}:  Up,
	{0, 1}:   Down,
	{-1, 0}:  Left,
	{1, 0}:   Right,
	{-1, -1}: UpLeft,
	{1, -1}:  UpRight,
	{-1, 1}:  DownLeft,
	{1, 1}:   DownRight,
}

var DirToVector = map[PlayerDir]Vec{
	Up:        {0, -1},
	Down:      {0, 1},
	Left:      {-1, 0},
	Right:     {1, 0},
	UpLeft:    {-1, -1},
	UpRight:   {1, -1},
	DownLeft:  {-1, 1},
	DownRight: {1, 1},
}
