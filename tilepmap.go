package main

import (
	"encoding/json"
	"os"
)

type TileJSON struct {
	Id string `json:"id"`
	X  int    `json:"x"`
	Y  int    `json:"y"`
}

type LayerJSON struct {
	Name  string      `json:"name"`
	Tiles []*TileJSON `json:"tiles"`
}

type TilemapJSON struct {
	TileSize int          `json:"tileSize"`
	Width    int          `json:"mapWidth"`
	Height   int          `json:"mapHeight"`
	Layers   []*LayerJSON `json:"layers"`
}

func NewTilemap(filepath string) (*TilemapJSON, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var tilemap TilemapJSON
	if err := json.Unmarshal(content, &tilemap); err != nil {
		return nil, err
	}

	return &tilemap, nil
}
