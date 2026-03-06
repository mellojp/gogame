package main

import "math"

type PlayerMode uint8

const (
	Idle PlayerMode = iota
	Run
	Roll
	Attack1
	Attack2
	Attack3
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
	attackStepX       float64
	attackStepY       float64
	attackMomentumX   float64
	attackMomentumY   float64
	attackLungeTicks  int
	rollCooldownTicks int
	comboStep         int
	comboQueued       bool
	queuedAttackDir   PlayerDir
	rollQueued        bool
	queuedRollDir     PlayerDir
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
	finishedMode := p.CurrMode

	p.CurrMode = Idle
	p.Animations[p.CurrMode][p.CurrDirection].Reset()
	p.rollStepX = 0
	p.rollStepY = 0
	p.attackStepX = 0
	p.attackStepY = 0
	p.attackMomentumX = 0
	p.attackMomentumY = 0
	p.attackLungeTicks = 0
	p.comboStep = 0
	p.comboQueued = false
	p.rollQueued = false

	if finishedMode == Roll {
		p.rollCooldownTicks = ROLL_POST_DELAY_TICKS
	}

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

func (p *Player) IsActionLocked() bool {
	return p.isAnimationLocked
}

func (p *Player) IsAttacking() bool {
	return p.isAnimationLocked && (p.CurrMode == Attack1 || p.CurrMode == Attack2 || p.CurrMode == Attack3)
}

func (p *Player) IsInvulnerable() bool {
	return p.IsRolling()
}

func (p *Player) TryStartRoll(dir PlayerDir) bool {
	if p.rollCooldownTicks > 0 {
		return false
	}

	if p.IsAttacking() {
		if p.isComboQueueWindowOpen() {
			p.rollQueued = true
			p.queuedRollDir = dir
			p.comboQueued = false
			return true
		}
		return false
	}

	if p.isAnimationLocked {
		return false
	}

	return p.startRoll(dir)
}

func (p *Player) startRoll(dir PlayerDir) bool {
	rollAnims, ok := p.Animations[Roll]
	if !ok {
		return false
	}

	rollAnim, ok := rollAnims[dir]
	if !ok {
		return false
	}

	p.isAnimationLocked = true
	p.CurrDirection = dir
	p.CurrMode = Roll
	rollAnim.Reset()
	p.rollQueued = false
	p.comboQueued = false
	p.comboStep = 0
	p.attackMomentumX = 0
	p.attackMomentumY = 0
	p.attackStepX = 0
	p.attackStepY = 0
	p.attackLungeTicks = 0

	v, ok := DirToVector[dir]
	if !ok {
		v = DirToVector[p.CurrDirection]
	}

	size := 1.0
	if v.X != 0 && v.Y != 0 {
		size = 1.4142
	}

	rollDurationTicks := p.currentActionDurationTicks(Roll)
	step := ROLL_DISTANCE / float64(rollDurationTicks)
	p.rollStepX = (float64(v.X) / size) * step
	p.rollStepY = (float64(v.Y) / size) * step

	return true
}

func (p *Player) ConsumeRollStep() (dx, dy float64, ok bool) {
	if !p.IsRolling() {
		return 0, 0, false
	}

	return p.rollStepX, p.rollStepY, true
}

func (p *Player) TryStartAttack(dir PlayerDir) bool {
	if p.IsAttacking() {
		if p.comboStep < 3 && p.isComboQueueWindowOpen() {
			p.comboQueued = true
			p.queuedAttackDir = dir
			p.rollQueued = false
			return true
		}
		return false
	}

	if p.isAnimationLocked {
		return false
	}

	return p.startComboAttack(1, dir)
}

func (p *Player) ConsumeAttackStep() (dx, dy float64, ok bool) {
	if !p.IsAttacking() {
		return 0, 0, false
	}

	dx = p.attackMomentumX
	dy = p.attackMomentumY

	if p.attackLungeTicks > 0 {
		dx += p.attackStepX
		dy += p.attackStepY
		p.attackLungeTicks--

		p.attackMomentumX += p.attackStepX * ATK_LUNGE_CARRY_GAIN
		p.attackMomentumY += p.attackStepY * ATK_LUNGE_CARRY_GAIN
	}

	p.attackMomentumX *= ATK_MOMENTUM_DECAY
	p.attackMomentumY *= ATK_MOMENTUM_DECAY

	if p.attackLungeTicks == 0 && math.Abs(dx) < 0.01 && math.Abs(dy) < 0.01 {
		return 0, 0, false
	}

	return dx, dy, true
}

func (p *Player) AdvanceQueuedAction() bool {
	if !p.IsAttacking() {
		return false
	}

	if p.rollQueued {
		return p.startRoll(p.queuedRollDir)
	}

	if !p.comboQueued || p.comboStep >= 3 {
		return false
	}

	nextStep := p.comboStep + 1
	nextDir := p.queuedAttackDir
	return p.startComboAttack(nextStep, nextDir)
}

func (p *Player) startComboAttack(step int, dir PlayerDir) bool {
	mode, ok := attackModeForStep(step)
	if !ok {
		return false
	}

	modeAnims, ok := p.Animations[mode]
	if !ok {
		return false
	}
	modeAnim, ok := modeAnims[dir]
	if !ok {
		return false
	}

	p.isAnimationLocked = true
	p.CurrDirection = dir
	p.CurrMode = mode
	modeAnim.Reset()
	p.comboStep = step
	p.comboQueued = false
	p.queuedAttackDir = dir
	p.rollQueued = false

	v, ok := DirToVector[dir]
	if !ok {
		v = DirToVector[p.CurrDirection]
	}

	size := 1.0
	if v.X != 0 && v.Y != 0 {
		size = 1.4142
	}

	lungeDist, lungeTicks := attackLungeConfig(step)
	lungeStep := lungeDist / float64(lungeTicks)
	p.attackStepX = (float64(v.X) / size) * lungeStep
	p.attackStepY = (float64(v.Y) / size) * lungeStep
	p.attackLungeTicks = lungeTicks
	p.attackMomentumX *= ATK_CHAIN_MOMENTUM_KEEP
	p.attackMomentumY *= ATK_CHAIN_MOMENTUM_KEEP

	return true
}

func attackModeForStep(step int) (PlayerMode, bool) {
	switch step {
	case 1:
		return Attack1, true
	case 2:
		return Attack2, true
	case 3:
		return Attack3, true
	default:
		return Idle, false
	}
}

func attackLungeConfig(step int) (distance float64, ticks int) {
	ticks = ATK_LUNGE_TICKS
	if ticks < 1 {
		ticks = 1
	}

	distance = ATK_LUNGE_DIST
	if step == 2 {
		distance = ATK_LUNGE_DIST * 0.9
	} else if step == 3 {
		distance = ATK_LUNGE_DIST * 1.35
	}

	return distance, ticks
}

func (p *Player) isComboQueueWindowOpen() bool {
	anim := p.CurrentAnimation()

	totalFrames := (anim.LastFrame - anim.FirstFrame) + 1
	if totalFrames <= 0 {
		return true
	}

	progressFrames := (anim.Frame() - anim.FirstFrame) + 1
	if progressFrames < 1 {
		progressFrames = 1
	}

	return progressFrames*100 >= totalFrames*COMBO_QUEUE_OPEN_PERCENT
}

func (p *Player) currentActionDurationTicks(mode PlayerMode) int {
	modeAnims, ok := p.Animations[mode]
	if !ok {
		return 1
	}

	anim, ok := modeAnims[p.CurrDirection]
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
