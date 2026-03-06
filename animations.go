package main

type Animation struct {
	FirstFrame int
	LastFrame  int
	Step       int
	// SpeedInTps is the frame interval in ticks (can be fractional).
	SpeedInTps   float32
	FrameCounter float32
	currFrame    int
	loops        bool
}

func NewAnimation(first, last, step int, speed float32, loops bool) *Animation {
	return &Animation{
		FirstFrame:   first,
		LastFrame:    last,
		Step:         step,
		SpeedInTps:   speed,
		FrameCounter: 0,
		currFrame:    first,
		loops:        loops,
	}
}

func (a *Animation) Frame() int {
	return a.currFrame
}

func (a *Animation) Update() bool {
	frameInterval := a.SpeedInTps
	if frameInterval <= 0 {
		frameInterval = 1
	}

	a.FrameCounter += 1.0
	for a.FrameCounter >= frameInterval {
		a.FrameCounter -= frameInterval
		a.currFrame += a.Step

		if a.currFrame > a.LastFrame {
			if a.loops {
				a.currFrame = a.FirstFrame
				continue
			}
			a.currFrame = a.LastFrame
			a.FrameCounter = 0
			return true
		}
	}
	return false
}

func (a *Animation) Reset() {
	a.currFrame = a.FirstFrame
	a.FrameCounter = 0
}

func (a *Animation) SetLoop(condition bool) {
	a.loops = condition
}
