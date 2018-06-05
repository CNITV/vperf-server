package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// MainTicker is the global ticker
var MainTicker *Ticker

func initTicker() {
	MainTicker = NewTicker(time.Duration(Conf.Time) * time.Minute)
}

// Ticker keeps the contest time. It can be paused and resumed
type Ticker struct {
	Running    bool
	Duration   time.Duration
	Prev       time.Duration
	LastMinute int

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
		t.LastMinute = int(t.ElapsedTime().Minutes())
	F:
		for {
			if t.RemainingTime() <= 0 {
				break
			}
			t.RecalculateProblemScore()
			select {
			case <-time.After(time.Second):
			case <-t.stopChan:
				break F
			}
		}
		t.Running = false
		t.Duration -= time.Since(t.startTime)
		t.Prev += time.Since(t.startTime)
		logrus.Info("Contest stopped")
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

func (t *Ticker) RecalculateProblemScore() {
	min := int(t.ElapsedTime().Minutes())
	if min > t.LastMinute {
		delta := min - t.LastMinute
		if t.RemainingTime().Minutes() < 20 {
			delta -= 20 - int(t.RemainingTime().Minutes())
		}
		if delta <= 0 {
			return
		}
		log := logrus.WithField("minutes", min)
		log.Infof("%d minutes passed. Going to increase problem scores", min-t.LastMinute)
		inc := []string{}
		for i, p := range Store.passed {
			// if the task is still unsolved
			if p == 0 {
				Store.Problems[i].Score += min - t.LastMinute
				inc = append(inc, strconv.Itoa(i))
			}
		}
		log.Infof("Tasks %s had their scores increased", strings.Join(inc, ", "))
		t.LastMinute = int(t.ElapsedTime().Minutes())
	}
}
