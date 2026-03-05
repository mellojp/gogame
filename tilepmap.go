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
	Name     string      `json:"name"`
	Tiles    []*TileJSON `json:"tiles"`
	Collider bool        `json:"collider"`
}

type TilemapJSON struct {
	TileSize int          `json:"tileSize"`
	Width    int          `json:"mapWidth"`
	Height   int          `json:"mapHeight"`
	Layers   []*LayerJSON `json:"layers"`
	Solid    [][]bool
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

	tilemap.Solid = make([][]bool, tilemap.Height)
	for y := 0; y < tilemap.Height; y++ {
		tilemap.Solid[y] = make([]bool, tilemap.Width)
	}

	for _, layer := range tilemap.Layers {
		if layer.Collider {
			for _, tile := range layer.Tiles {
				if tile.X >= 0 && tile.X < tilemap.Width && tile.Y >= 0 && tile.Y < tilemap.Height {
					tilemap.Solid[tile.Y][tile.X] = true
				}
			}
		}
	}

	return &tilemap, nil
}

func (t *TilemapJSON) IsTileSolid(tx, ty int) bool {
	if tx < 0 || ty < 0 || tx >= t.Width || ty >= t.Height {
		return true
	}
	return t.Solid[ty][tx]
}
