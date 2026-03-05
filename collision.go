package main

import "math"

func (g *Game) movePlayerWithCollision(inputX, inputY int) {
	oldX, oldY := g.Player.Xpos, g.Player.Ypos
	g.Player.Move(inputX, inputY)
	dx, dy := g.Player.Dx, g.Player.Dy
	g.Player.Xpos, g.Player.Ypos = oldX, oldY

	g.movePlayerByVelocityWithCollision(dx, dy)
}

func (g *Game) movePlayerByVelocityWithCollision(dx, dy float64) {
	g.movePlayerAxis(dx, 0)
	g.movePlayerAxis(0, dy)
}

func (g *Game) movePlayerAxis(dx, dy float64) {
	steps := int(math.Ceil(math.Max(math.Abs(dx), math.Abs(dy))))
	if steps <= 0 {
		return
	}

	stepX := dx / float64(steps)
	stepY := dy / float64(steps)

	for i := 0; i < steps; i++ {
		nextX := g.Player.Xpos + stepX
		nextY := g.Player.Ypos + stepY
		if g.playerCollidesAt(nextX, nextY) {
			return
		}
		g.Player.Xpos = nextX
		g.Player.Ypos = nextY
	}
}

func (g *Game) playerCollidesAt(playerX, playerY float64) bool {
	left, top, right, bottom := g.Player.HitboxAt(playerX, playerY)
	tileSize := float64(g.Tilemap.TileSize) * MAP_DRAW_SCALE

	minTX := int(math.Floor(left / tileSize))
	maxTX := int(math.Floor((right - 1) / tileSize))
	minTY := int(math.Floor(top / tileSize))
	maxTY := int(math.Floor((bottom - 1) / tileSize))

	for ty := minTY; ty <= maxTY; ty++ {
		for tx := minTX; tx <= maxTX; tx++ {
			if g.Tilemap.IsTileSolid(tx, ty) {
				return true
			}
		}
	}

	return false
}
