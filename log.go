package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Duration is a magic time.Duration that can be marshalled into JSON
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

// Log is the global Event Log
var Log *EventLog

// EventLog is the chain of events that happen during the contest
type EventLog struct {
	Entries []LogEntry
	path    string
	m       *sync.Mutex
}

// LogEvent is a type of event that happens during the contest
type LogEvent string

const (
	// EventSetSpecial has two params: team_id and task_id
	EventSetSpecial LogEvent = "SET_SPECIAL"
	// EventSubmitAnswer has three params: team_id, task_id and answer
	EventSubmitAnswer LogEvent = "SUBMIT_ANSWER"
	// EventFineTeam has two params: team_id and points
	EventFineTeam LogEvent = "FINE_TEAM"
	// EventDisqualifyTeam has one param: team_id
	EventDisqualifyTeam LogEvent = "DISQUALIFY_TEAM"
)

// LogEntry has the complete details of an event that happens during the contest
type LogEntry struct {
	Event          LogEvent       `json:"event"`
	Params         map[string]int `json:"params"`
	TimeSinceStart Duration       `json:"time_since_start"`
}

// Save saves the events in a JSON file
func (el *EventLog) Save() {
	j, err := json.MarshalIndent(el.Entries, "", "  ")
	if err != nil {
		logrus.Panic(err)
		return
	}

	f, err := os.Create(el.path)
	if err != nil {
		logrus.WithField("path", el.path).Error("Couldn't open file to save log. Dumping JSON to stdout")
		fmt.Println(string(j))
	}
	defer f.Close()
	fmt.Fprintln(f, string(j))
}

// Process loads the events in the state
func (el *EventLog) Process() {
	logrus.Info("Processing previous log")
	for _, e := range el.Entries {
		MainTicker.Prev = e.TimeSinceStart.Duration
		switch e.Event {
		case EventSetSpecial:
			setSpecial(e.Params["team_id"], e.Params["task_id"])
		case EventSubmitAnswer:
			MainTicker.RecalculateProblemScore()
			submitAnswer(e.Params["team_id"], e.Params["task_id"], e.Params["answer"])
		case EventFineTeam:
			fineTeam(e.Params["team_id"], e.Params["points"])
		case EventDisqualifyTeam:
			disqualifyTeam(e.Params["team_id"])
		}
	}
	MainTicker.RecalculateProblemScore()
}

func (el *EventLog) Push(ev LogEvent, params map[string]int) {
	el.m.Lock()
	defer el.m.Unlock()

	el.Entries = append(el.Entries, LogEntry{
		Event:          ev,
		Params:         params,
		TimeSinceStart: Duration{Duration: MainTicker.ElapsedTime()},
	})
	logrus.Debug(el.Entries)
}

// Delete deletes an entry
func (el *EventLog) Delete(is ...int) {
	el.m.Lock()
	defer el.m.Unlock()
	new := []LogEntry{}
	for ei := range el.Entries {
		// if this entry is not in the parameter list
		i := 0
		for i < len(is) && ei != is[i] {
			i++
		}
		// keep it
		if i == len(is) {
			new = append(new, el.Entries[ei])
		}
	}
	el.Entries = new
}

func initLog(path string) {
	Log = &EventLog{path: path, m: &sync.Mutex{}}

	log := logrus.WithField("path", path)
	log.Info("Reading log file")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("Log file doesn't exist, will be created. Starting from scratch...")
			return
		}
		log.WithError(err).Fatal("Log file cannot be opened")
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.WithError(err).Fatal("Couldn't read log file")
	}
	err = json.Unmarshal(b, &Log.Entries)
	if err != nil {
		log.WithError(err).Fatal("Couldn't decode log file")
	}
	Log.Process()
	logrus.Warn("You need to start the contest now")
}
