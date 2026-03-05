package main

import (
	"image"
	"image/color"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	WINDOW_WIDTH  = 1280
	WINDOW_HEIGHT = 720

	VERTICAL_TILES    = 8
	HORTIZONTAL_TILES = 15
	TILE_SIZE         = 64

	MAP_DRAW_SCALE    = 4.0
	PLAYER_DRAW_SCALE = 4.0

	PLAYER_SPEED = 5.5
	RUN_ANIM_TPS = 3.0

	IDLE_ANIM_TPS = 6.0

	ROLL_ANIM_TPS       = 2.0
	ROLL_DISTANCE       = 420.0
	ROLL_COOLDOWN_TICKS = 28

	DEBUG_PLAYER_HITBOX = false
)

type Game struct {
	Player *Player

	Tilemap    *TilemapJSON
	TilemapImg *ebiten.Image

	Cam *Camera
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

	g.Player.TickRollCooldown()
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		rollDir := g.Player.CurrDirection
		if inputX != 0 || inputY != 0 {
			if dir, ok := VectorToDir[Vec{X: inputX, Y: inputY}]; ok {
				rollDir = dir
			}
		}
		g.Player.TryStartRoll(rollDir)
	}

	if g.Player.IsRolling() {
		if rollDx, rollDy, ok := g.Player.ConsumeRollStep(); ok {
			g.movePlayerByVelocityWithCollision(rollDx, rollDy)
		}
	} else {
		g.Player.SetDirection(inputX, inputY)
		g.Player.SetMode(inputX, inputY)
		g.movePlayerWithCollision(inputX, inputY)
	}

	animFinished := g.Player.CurrentAnimation().Update()

	if g.Player.isAnimationLocked && animFinished {
		// Transition on the same tick to avoid an extra end-turn display frame.
		g.Player.FinishAction()
	}

	//camera settings
	spriteHalf := float64(g.Player.Modes[g.Player.CurrMode].TileSize) * PLAYER_DRAW_SCALE / 2.0
	g.Cam.FollowTarget(
		g.Player.Xpos+spriteHalf,
		g.Player.Ypos+spriteHalf,
		WINDOW_WIDTH,
		WINDOW_HEIGHT,
	)
	g.Cam.Constraint(
		float64(g.Tilemap.Width*g.Tilemap.TileSize)*MAP_DRAW_SCALE,
		float64(g.Tilemap.Height*g.Tilemap.TileSize)*MAP_DRAW_SCALE,
		WINDOW_WIDTH,
		WINDOW_HEIGHT,
	)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	//set player position
	opts := ebiten.DrawImageOptions{}
	opts.Filter = ebiten.FilterNearest
	opts.GeoM.Scale(PLAYER_DRAW_SCALE, PLAYER_DRAW_SCALE)
	opts.GeoM.Translate(g.Player.Xpos, g.Player.Ypos)
	opts.GeoM.Translate(g.Cam.X, g.Cam.Y)

	//loop over the tilemap layers
	tilesetColumns := g.TilemapImg.Bounds().Dx() / g.Tilemap.TileSize
	for i := len(g.Tilemap.Layers) - 1; i >= 0; i-- {
		layer := g.Tilemap.Layers[i]
		for _, tile := range layer.Tiles {
			tileOpts := ebiten.DrawImageOptions{}
			tileOpts.Filter = ebiten.FilterNearest

			x := float64(tile.X * g.Tilemap.TileSize)
			y := float64(tile.Y * g.Tilemap.TileSize)
			tileOpts.GeoM.Scale(MAP_DRAW_SCALE, MAP_DRAW_SCALE)
			tileOpts.GeoM.Translate(x*MAP_DRAW_SCALE, y*MAP_DRAW_SCALE)
			tileOpts.GeoM.Translate(g.Cam.X, g.Cam.Y)
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

	if DEBUG_PLAYER_HITBOX {
		left, top, right, bottom := g.Player.HitboxAt(g.Player.Xpos, g.Player.Ypos)
		hitboxX := left + g.Cam.X
		hitboxY := top + g.Cam.Y
		hitboxW := right - left
		hitboxH := bottom - top

		ebitenutil.DrawRect(screen, hitboxX, hitboxY, hitboxW, hitboxH, color.RGBA{R: 255, G: 0, B: 0, A: 90})

		const border = 2.0
		ebitenutil.DrawRect(screen, hitboxX, hitboxY, hitboxW, border, color.RGBA{R: 255, G: 255, B: 255, A: 220})
		ebitenutil.DrawRect(screen, hitboxX, hitboxY+hitboxH-border, hitboxW, border, color.RGBA{R: 255, G: 255, B: 255, A: 220})
		ebitenutil.DrawRect(screen, hitboxX, hitboxY, border, hitboxH, color.RGBA{R: 255, G: 255, B: 255, A: 220})
		ebitenutil.DrawRect(screen, hitboxX+hitboxW-border, hitboxY, border, hitboxH, color.RGBA{R: 255, G: 255, B: 255, A: 220})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WINDOW_WIDTH, WINDOW_HEIGHT
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
	idleImg, _, err := ebitenutil.NewImageFromFile("assets/player/Idle.png")
	if err != nil {
		log.Fatal(err)
	}
	runImg, _, err := ebitenutil.NewImageFromFile("assets/player/Run.png")
	if err != nil {
		log.Fatal(err)
	}
	rollImg, _, err := ebitenutil.NewImageFromFile("assets/player/Rolling.png")
	if err != nil {
		log.Fatal(err)
	}
	//maps loading
	mapImg, _, err := ebitenutil.NewImageFromFile("assets/maps/spritesheet2.png")
	if err != nil {
		log.Fatal(err)
	}
	tilemap, err := NewTilemap("assets/maps/map2.json")
	if err != nil {
		log.Fatal(err)
	}

	playerSprisheets := map[PlayerMode]*Spritesheet{
		Idle: NewSpritesheet(VERTICAL_TILES, HORTIZONTAL_TILES, TILE_SIZE, idleImg),
		Run:  NewSpritesheet(VERTICAL_TILES, HORTIZONTAL_TILES, TILE_SIZE, runImg),
		Roll: NewSpritesheet(VERTICAL_TILES, HORTIZONTAL_TILES, TILE_SIZE, rollImg),
	}

	playerAnimations := map[PlayerMode]map[PlayerDir]*Animation{
		Idle: newDirectionalAnimations(IDLE_ANIM_TPS, true),
		Run:  newDirectionalAnimations(RUN_ANIM_TPS, true),
		Roll: newDirectionalAnimations(ROLL_ANIM_TPS, false),
	}

	player := NewPlayer(
		(float64(tilemap.Width*tilemap.TileSize)*MAP_DRAW_SCALE-float64(TILE_SIZE)*PLAYER_DRAW_SCALE)/2.0,
		(float64(tilemap.Height*tilemap.TileSize)*MAP_DRAW_SCALE-float64(TILE_SIZE)*PLAYER_DRAW_SCALE)/2.0,
		playerAnimations,
		playerSprisheets,
	)

	if err := ebiten.RunGame(
		&Game{
			Player:     player,
			Tilemap:    tilemap,
			TilemapImg: mapImg,
			Cam:        NewCamera(0, 0),
		},
	); err != nil {
		log.Fatal(err)
	}
}
