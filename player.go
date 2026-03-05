package main

type PlayerMode uint8

const (
	Idle PlayerMode = iota
	Run
	Roll
)

type Player struct {
	Xpos, Ypos float64
	Dx, Dy     float64

	Animations        map[PlayerMode]map[PlayerDir]*Animation
	isAnimationLocked bool
	CurrDirection     PlayerDir
	Modes             map[PlayerMode]*Spritesheet
	CurrMode          PlayerMode

	HitboxOffsetX float64
	HitboxOffsetY float64
	HitboxWidth   float64
	HitboxHeight  float64

	rollStepX         float64
	rollStepY         float64
	rollCooldownTicks int
}

func NewPlayer(x, y float64, animations map[PlayerMode]map[PlayerDir]*Animation, modes map[PlayerMode]*Spritesheet) *Player {
	return &Player{
		Xpos:          x,
		Ypos:          y,
		Animations:    animations,
		CurrDirection: DownLeft,
		Modes:         modes,
		CurrMode:      Idle,
		// Feet-centered hitbox in world pixels (scaled with player render scale).
		HitboxOffsetX: 26.0 * PLAYER_DRAW_SCALE,
		HitboxOffsetY: 30 * PLAYER_DRAW_SCALE,
		HitboxWidth:   12.0 * PLAYER_DRAW_SCALE,
		HitboxHeight:  14.0 * PLAYER_DRAW_SCALE,
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
	p.rollStepX = 0
	p.rollStepY = 0

	p.isAnimationLocked = false
}

func (p *Player) TickRollCooldown() {
	if p.rollCooldownTicks > 0 {
		p.rollCooldownTicks--
	}
}

func (p *Player) IsRolling() bool {
	return p.isAnimationLocked && p.CurrMode == Roll
}

func (p *Player) IsInvulnerable() bool {
	return p.IsRolling()
}

func (p *Player) TryStartRoll(dir PlayerDir) bool {
	if p.isAnimationLocked || p.rollCooldownTicks > 0 {
		return false
	}

	p.CurrDirection = dir
	p.TriggerAction(Roll)
	if !p.IsRolling() {
		return false
	}

	v, ok := DirToVector[dir]
	if !ok {
		v = DirToVector[p.CurrDirection]
	}

	size := 1.0
	if v.X != 0 && v.Y != 0 {
		size = 1.4142
	}

	rollDurationTicks := p.currentRollDurationTicks()
	step := ROLL_DISTANCE / float64(rollDurationTicks)
	p.rollStepX = (float64(v.X) / size) * step
	p.rollStepY = (float64(v.Y) / size) * step
	p.rollCooldownTicks = ROLL_COOLDOWN_TICKS

	return true
}

func (p *Player) ConsumeRollStep() (dx, dy float64, ok bool) {
	if !p.IsRolling() {
		return 0, 0, false
	}

	return p.rollStepX, p.rollStepY, true
}

func (p *Player) currentRollDurationTicks() int {
	rollAnims, ok := p.Animations[Roll]
	if !ok {
		return 1
	}

	anim, ok := rollAnims[p.CurrDirection]
	if !ok {
		return 1
	}

	if anim.Step <= 0 || anim.LastFrame < anim.FirstFrame {
		return 1
	}

	framesToAdvance := ((anim.LastFrame - anim.FirstFrame) / anim.Step) + 1
	if framesToAdvance <= 1 {
		return 1
	}

	// Animation.Update() decrementa o contador em 1 por tick.
	// Para speed positivo, cada avanço de frame ocorre a cada int(speed)+1 ticks.
	advanceIntervalTicks := int(anim.SpeedInTps) + 1
	if advanceIntervalTicks < 1 {
		advanceIntervalTicks = 1
	}

	return 1 + (framesToAdvance-1)*advanceIntervalTicks
}

func (p *Player) HitboxAt(x, y float64) (left, top, right, bottom float64) {
	left = x + p.HitboxOffsetX
	top = y + p.HitboxOffsetY
	right = left + p.HitboxWidth
	bottom = top + p.HitboxHeight

	if p.IsRolling() {
		// Slightly larger during roll, with extra reach toward roll direction.
		baseExpand := 2.0 * PLAYER_DRAW_SCALE
		forwardExpand := 4.0 * PLAYER_DRAW_SCALE

		left -= baseExpand
		top -= baseExpand
		right += baseExpand
		bottom += baseExpand

		if v, ok := DirToVector[p.CurrDirection]; ok {
			if v.X < 0 {
				left -= forwardExpand
			} else if v.X > 0 {
				right += forwardExpand
			}

			if v.Y < 0 {
				top -= forwardExpand
			} else if v.Y > 0 {
				bottom += forwardExpand
			}
		}
	}

	return
}
