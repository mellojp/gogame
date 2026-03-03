package main

type Animation struct {
	FirstFrame   int
	LastFrame    int
	Step         int
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
	a.FrameCounter -= 1.0
	if a.FrameCounter < 0.0 {
		a.FrameCounter = a.SpeedInTps
		a.currFrame += a.Step

		if a.currFrame > a.LastFrame {
			if a.loops {
				a.Reset()
			} else {
				a.currFrame = a.LastFrame
				return true
			}

		}
	}
	return false
}

func (a *Animation) Reset() {
	a.currFrame = a.FirstFrame
}

func (a *Animation) SetLoop(condition bool) {
	a.loops = condition
}
