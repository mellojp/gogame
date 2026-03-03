package main

type PlayerMode uint8

const (
	Idle PlayerMode = iota
	Run
	Stop
)

type Player struct {
	Xpos, Ypos        float64
	Dx, Dy            float64
	Animations        map[PlayerMode]map[PlayerDir]*Animation
	CurrDirection     PlayerDir
	Modes             map[PlayerMode]*Spritesheet
	CurrMode          PlayerMode
	isAnimationLocked bool
}

func NewPlayer(x, y float64, animations map[PlayerMode]map[PlayerDir]*Animation, modes map[PlayerMode]*Spritesheet) *Player {
	return &Player{
		Xpos:          x,
		Ypos:          y,
		Animations:    animations,
		CurrDirection: DownLeft,
		Modes:         modes,
		CurrMode:      Idle,
	}
}

func (p *Player) Move(x, y int) {
	size := 0.0
	if x != 0 && y != 0 {
		size = 1.4142 // Movimento diagonal constante
	} else if x != 0 || y != 0 {
		size = 1.0 // Movimento reto
	}

	if size > 0 {
		speed := float64(PLAYER_SPEED)
		if p.isAnimationLocked {
			speed *= 0.3
		}
		p.Dx = (float64(x) / size) * speed
		p.Dy = (float64(y) / size) * speed
	}

	p.Xpos += p.Dx
	p.Ypos += p.Dy
}

func (p *Player) SetDirection(x, y int) {
	if p.isAnimationLocked {
		return
	}

	if x == 0 && y == 0 {
		return
	}

	v := Vec{
		X: x,
		Y: y,
	}

	dir, ok := VectorToDir[v]
	if !ok {
		return
	}

	// APAGUE ESTE BLOCO QUE EXISTIA AQUI:
	// current := DirToVector[p.CurrDirection]
	// if current.X == -x && current.Y == -y {
	// 	p.TriggerAction(Turn)
	// 	return
	// }

	if p.CurrDirection == dir {
		return
	}

	p.CurrDirection = dir
	p.Animations[p.CurrMode][p.CurrDirection].Reset()
}

func (p *Player) SetMode(x, y int) {
	if p.isAnimationLocked {
		return
	}

	next := Idle
	if x != 0 || y != 0 {
		next = Run
	}

	if next != p.CurrMode {
		p.CurrMode = next
		p.Animations[p.CurrMode][p.CurrDirection].Reset()
	}
}

func (p *Player) TriggerAction(m PlayerMode) {
	if !p.isAnimationLocked {
		if _, ok := p.Animations[m]; ok { // Verifica se o modo existe
			p.isAnimationLocked = true
			p.Animations[m][p.CurrDirection].Reset()
			p.CurrMode = m
		}
	}
}

func (p *Player) CurrentAnimation() *Animation {
	if anims, ok := p.Animations[p.CurrMode]; ok {
		if anim, ok := anims[p.CurrDirection]; ok {
			return anim
		}
	}
	// Retorna animação padrão (Idle) se der problema, evitando crash
	return p.Animations[Idle][p.CurrDirection]
}
func (p *Player) FinishAction() {

	p.CurrMode = Idle
	p.Animations[p.CurrMode][p.CurrDirection].Reset()

	p.isAnimationLocked = false
}
