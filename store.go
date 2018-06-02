package main

var Store *Storage

type storeTeam struct {
	ID      int          `json:"id"`
	Name    string       `json:"name"`
	Score   int          `json:"score"`
	Special int          `json:"special"`
	Trials  []storeTrial `json:"trials"`
}

type storeTrial struct {
	No     int  `json:"no"`
	Passed bool `json:"passed"`
}

type storeProblem struct {
	ID    int `json:"id"`
	Score int `json:"score"`
}

type Storage struct {
	// Time is in seconds
	Time        int            `json:"time"`
	TotalTime   int            `json:"total_time"`
	Running     bool           `json:"running"`
	PauseReason string         `json:"pause_reason"`
	Teams       []storeTeam    `json:"teams"`
	Problems    []storeProblem `json:"problems"`
}

func initStorage(c *Config) {
	s := &Storage{
		Time:      c.Time * 60,
		TotalTime: c.Time * 60,
		Teams:     []storeTeam{},
		Problems:  []storeProblem{},
	}

	for i, t := range c.Teams {
		team := storeTeam{
			ID:      i,
			Name:    t.Name,
			Score:   c.DefaultTeamScore,
			Special: 0,
			Trials:  []storeTrial{},
		}
		for i := 0; i < len(c.Solutions); i++ {
			team.Trials = append(team.Trials, storeTrial{No: 0, Passed: false})
		}
		s.Teams = append(s.Teams, team)
	}

	for i := 0; i < len(c.Solutions); i++ {
		s.Problems = append(s.Problems, storeProblem{
			ID:    i,
			Score: c.DefaultProblemScore,
		})
	}

	Store = s
}
