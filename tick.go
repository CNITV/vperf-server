package main

import "time"

var MainTicker *Ticker

func initTicker() {
	MainTicker = NewTicker(time.Duration(Conf.Time) * time.Minute)
}

type Ticker struct {
	Running  bool
	Duration time.Duration
	Prev     time.Duration

	startTime time.Time
	stopChan  chan bool
}

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

func (t *Ticker) Stop() {
	if !t.Running {
		return
	}
	t.stopChan <- true
}

func (t *Ticker) ElapsedSinceStart() time.Duration {
	if !t.Running {
		return 0
	}
	return time.Since(t.startTime)
}
