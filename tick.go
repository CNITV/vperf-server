package main

import "time"

// MainTicker is the global ticker
var MainTicker *Ticker

func initTicker() {
	MainTicker = NewTicker(time.Duration(Conf.Time) * time.Minute)
}

// Ticker keeps the contest time. It can be paused and resumed
type Ticker struct {
	Running  bool
	Duration time.Duration
	Prev     time.Duration

	startTime time.Time
	stopChan  chan bool
}

// NewTicker returns a new ticker with a base duration
func NewTicker(d time.Duration) *Ticker {
	t := &Ticker{
		Running:  false,
		Duration: d,
		Prev:     0,

		startTime: time.Time{},
		stopChan:  make(chan bool),
	}

	return t
}

// Start starts ticking until the time runs out or until the ticker is stopped
func (t *Ticker) Start() {
	t.startTime = time.Now()
	t.Running = true
	go func() {
	F:
		for {
			if time.Since(t.startTime) >= t.Duration {
				break
			}
			select {
			case <-time.After(time.Second):
			case <-t.stopChan:
				break F
			}
		}
		t.Running = false
		t.Duration -= time.Since(t.startTime)
		t.Prev += time.Since(t.startTime)
	}()
}

// Stop stops the ticker. If it didn't run out, it can be resumed later.
func (t *Ticker) Stop() {
	if !t.Running {
		return
	}
	t.stopChan <- true
}

// ElapsedSinceStart returns the amount of time elapsed since the ticker was resumed (or first start)
func (t *Ticker) ElapsedSinceResume() time.Duration {
	if !t.Running {
		return 0
	}
	return time.Since(t.startTime)
}

// RemainingTime returns the time left until the end of the contest
func (t *Ticker) RemainingTime() time.Duration {
	return time.Duration(Conf.Time)*time.Minute - t.ElapsedTime()
}

// ElapsedTime returns the elapsed contest time, not counting and break
func (t *Ticker) ElapsedTime() time.Duration {
	return MainTicker.Prev + t.ElapsedSinceResume()
}
