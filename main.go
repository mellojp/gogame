package main

import (
	"image"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	WINDOW_WIDTH  = 800
	WINDOW_HEIGHT = 600

	VERTICAL_TILES    = 8
	HORTIZONTAL_TILES = 15
	TILE_SIZE         = 64

	PLAYER_SPEED = 4

	IDLE_ANIM_TPS = 6.0
	RUN_ANIM_TPS  = 3.0
	STOP_ANIM_TPS = 0.5
)

type Game struct {
	Player *Player

	Tilemap    *TilemapJSON
	TilemapImg *ebiten.Image
}

func (g *Game) Update() error {
	g.Player.Dx = 0.0
	g.Player.Dy = 0.0

	inputX, inputY := 0, 0

	//move player
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		inputX += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		inputX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		inputY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		inputY += 1
	}

	g.Player.SetDirection(inputX, inputY)
	g.Player.SetMode(inputX, inputY)
	g.Player.Move(inputX, inputY)

	animFinished := g.Player.CurrentAnimation().Update()

	if g.Player.isAnimationLocked && animFinished {
		// Transition on the same tick to avoid an extra end-turn display frame.
		g.Player.FinishAction()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	//set player position
	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Scale(5, 5)
	opts.GeoM.Translate(g.Player.Xpos, g.Player.Ypos)

	//loop over the tilemap layers
	tilesetColumns := g.TilemapImg.Bounds().Dx() / g.Tilemap.TileSize
	for i := len(g.Tilemap.Layers) - 1; i >= 0; i-- {
		layer := g.Tilemap.Layers[i]
		for _, tile := range layer.Tiles {
			tileOpts := ebiten.DrawImageOptions{}

			x := float64(tile.X * g.Tilemap.TileSize)
			y := float64(tile.Y * g.Tilemap.TileSize)
			tileOpts.GeoM.Translate(x, y)

			id, _ := strconv.Atoi(tile.Id)

			srcX := (id % tilesetColumns) * g.Tilemap.TileSize
			srcY := (id / tilesetColumns) * g.Tilemap.TileSize

			rect := image.Rect(
				srcX,
				srcY,
				srcX+g.Tilemap.TileSize,
				srcY+g.Tilemap.TileSize,
			)

			screen.DrawImage(
				g.TilemapImg.SubImage(rect).(*ebiten.Image),
				&tileOpts,
			)

		}
	}

	//draw player
	screen.DrawImage(
		//chop spritesheet
		g.Player.Modes[g.Player.CurrMode].Image.SubImage(
			g.Player.Modes[g.Player.CurrMode].Chop(g.Player.CurrentAnimation().Frame()),
		).(*ebiten.Image),
		&opts,
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ebiten.WindowSize()
}

func newDirectionalAnimations(speed float32, loops bool) map[PlayerDir]*Animation {
	return map[PlayerDir]*Animation{
		Right:     NewAnimation(0, 14, 1, speed, loops),
		DownRight: NewAnimation(15, 29, 1, speed, loops),
		Down:      NewAnimation(30, 44, 1, speed, loops),
		DownLeft:  NewAnimation(45, 59, 1, speed, loops),
		Left:      NewAnimation(60, 74, 1, speed, loops),
		UpLeft:    NewAnimation(75, 89, 1, speed, loops),
		Up:        NewAnimation(90, 104, 1, speed, loops),
		UpRight:   NewAnimation(105, 119, 1, speed, loops),
	}
}

func main() {
	ebiten.SetWindowSize(WINDOW_WIDTH, WINDOW_HEIGHT)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	//player imgs loading
	idleImg, _, err := ebitenutil.NewImageFromFile("assets/player/Shadowless/Idle2_Shadowless.png")
	if err != nil {
		log.Fatal(err)
	}
	runImg, _, err := ebitenutil.NewImageFromFile("assets/player/Shadowless/Run_Shadowless.png")
	if err != nil {
		log.Fatal(err)
	}
	stopImg, _, err := ebitenutil.NewImageFromFile("assets/player/Shadowless/Stop_Shadowless.png")
	if err != nil {
		log.Fatal(err)
	}
	//maps loading
	mapImg, _, err := ebitenutil.NewImageFromFile("assets/maps/spritesheet.png")
	if err != nil {
		log.Fatal(err)
	}
	tilemap, err := NewTilemap("assets/maps/map.json")
	if err != nil {
		log.Fatal(err)
	}

	playerSprisheets := map[PlayerMode]*Spritesheet{
		Idle: NewSpritesheet(VERTICAL_TILES, HORTIZONTAL_TILES, TILE_SIZE, idleImg),
		Run:  NewSpritesheet(VERTICAL_TILES, HORTIZONTAL_TILES, TILE_SIZE, runImg),
		Stop: NewSpritesheet(VERTICAL_TILES, HORTIZONTAL_TILES, TILE_SIZE, stopImg),
	}

	playerAnimations := map[PlayerMode]map[PlayerDir]*Animation{
		Idle: newDirectionalAnimations(IDLE_ANIM_TPS, true),
		Run:  newDirectionalAnimations(RUN_ANIM_TPS, true),
		Stop: newDirectionalAnimations(STOP_ANIM_TPS, false),
	}

	player := NewPlayer(
		(WINDOW_WIDTH / 2),
		(WINDOW_HEIGHT / 2),
		playerAnimations,
		playerSprisheets,
	)

	if err := ebiten.RunGame(
		&Game{
			Player:     player,
			Tilemap:    tilemap,
			TilemapImg: mapImg,
		},
	); err != nil {
		log.Fatal(err)
	}
}
