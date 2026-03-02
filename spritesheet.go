package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Spritesheet struct {
	Image           *ebiten.Image
	VerticalTiles   int
	HorizontalTiles int
	TileSize        int
}

func NewSpritesheet(vt, ht, ts int, image *ebiten.Image) *Spritesheet {
	return &Spritesheet{
		Image:           image,
		VerticalTiles:   vt,
		HorizontalTiles: ht,
		TileSize:        ts,
	}
}

func (s *Spritesheet) Chop(index int) image.Rectangle {
	x := (index % s.HorizontalTiles) * s.TileSize
	y := (index / s.HorizontalTiles) * s.TileSize

	return image.Rect(x, y, x+s.TileSize, y+s.TileSize)
}
