package main

type PlayerMode uint8

const (
	Idle PlayerMode = iota
	Run
	Turn
	Stop
)

type Player struct {
	Xpos, Ypos        float64
	Dx, Dy            float64
	Animations        map[PlayerMode]map[PlayerDir]*Animation
	CurrDirection     PlayerDir
	turnTargetDir     PlayerDir
	hasTurnTarget     bool
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

	current := DirToVector[p.CurrDirection]
	if current.X == -x && current.Y == -y {
		p.turnTargetDir = dir
		p.hasTurnTarget = true
		p.TriggerAction(Turn)
		return
	}

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

	if p.CurrMode == Run && next == Idle {
		p.TriggerAction(Stop)
		return
	}

	if next != p.CurrMode {
		p.CurrMode = next
		p.Animations[p.CurrMode][p.CurrDirection].Reset()
	}
}

func (p *Player) TriggerAction(m PlayerMode) {
	if !p.isAnimationLocked {
		p.isAnimationLocked = true
		p.Animations[m][p.CurrDirection].Reset()
		p.CurrMode = m
	}
}

func (p *Player) CurrentAnimation() *Animation {
	return p.Animations[p.CurrMode][p.CurrDirection]
}

func (p *Player) FinishAction() {
	if p.CurrMode == Turn && p.hasTurnTarget {
		p.CurrDirection = p.turnTargetDir
		p.hasTurnTarget = false
	}

	p.CurrMode = Idle
	p.Animations[p.CurrMode][p.CurrDirection].Reset()

	p.isAnimationLocked = false
}
