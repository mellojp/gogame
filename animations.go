package main

type Animation struct {
	FirstFrame   int
	LastFrame    int
	Step         int
	SpeedInTps   float32
	FrameCounter float32
	currFrame    int
	finished     bool
	loops        bool
	rangeFirst   int
	rangeLast    int
	hasRange     bool
}

func NewAnimation(first, last, step int, speed float32, loops bool) *Animation {
	return &Animation{
		FirstFrame:   first,
		LastFrame:    last,
		Step:         step,
		SpeedInTps:   speed,
		FrameCounter: 0,
		currFrame:    first,
		finished:     false,
		loops:        loops,
	}
}

func (a *Animation) Frame() int {
	return a.currFrame
}

func (a *Animation) Update() bool {
	if a.finished {
		return true
	}

	first := a.FirstFrame
	last := a.LastFrame
	if a.hasRange {
		first = a.rangeFirst
		last = a.rangeLast
	}

	if a.SpeedInTps <= 0 {
		a.SpeedInTps = 1
	}

	a.FrameCounter += 1.0
	if a.FrameCounter < a.SpeedInTps {
		return false
	}

	a.FrameCounter = 0
	a.currFrame += a.Step

	pastEnd := (a.Step >= 0 && a.currFrame > last) || (a.Step < 0 && a.currFrame < last)
	if pastEnd {
		if a.loops {
			a.currFrame = first
		} else {
			a.currFrame = last
			a.finished = true
			return true
		}
	}

	return false
}

func (a *Animation) Reset() {
	if a.hasRange {
		a.currFrame = a.rangeFirst
	} else {
		a.currFrame = a.FirstFrame
	}
	a.FrameCounter = 0
	a.finished = false
}

func (a *Animation) SetLoop(condition bool) {
	a.loops = condition
	if condition {
		a.finished = false
	}
}

func (a *Animation) SetRange(first, last int) {
	if a.Step >= 0 && first > last {
		first, last = last, first
	}
	if a.Step < 0 && first < last {
		first, last = last, first
	}
	a.rangeFirst = first
	a.rangeLast = last
	a.hasRange = true
}

func (a *Animation) ClearRange() {
	a.hasRange = false
}
