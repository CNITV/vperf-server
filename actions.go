package main

import "github.com/sirupsen/logrus"

func setSpecial(i, p int) {
	log := logrus.WithFields(logrus.Fields{
		"team": i,
		"task": p,
	})
	log.Info("Team set special task")
	if Store.Teams[i].setSpecial {
		log.Info("Team already set their special problem. Ignoring.")
		return
	}
	Store.Teams[i].Special = p
	Store.Teams[i].setSpecial = true
}

func submitAnswer(i, p, ans int) {
	if Store.Teams[i].Trials[p].Passed {
		logrus.WithFields(logrus.Fields{
			"team": i,
			"task": p,
		}).Info("Ignored answer because team already completed task")
		return
	}
	delta := 0
	log := logrus.WithFields(logrus.Fields{
		"team":   i,
		"task":   p,
		"answer": ans,
	})
	log.Info("Team submitted answer")
	if Conf.Solutions[p] != ans {
		// incorrect answer, remove 10 points
		delta -= 10
		Store.Teams[i].Trials[p].Passed = false
		Store.Problems[p].Score += 2
		log.Info("Team supplied wrong answer")
	} else {
		// correct answer, give points
		log.Infof("Team supplied good answer. Awarding %d points", Store.Problems[p].Score)
		delta += Store.Problems[p].Score
		// give bonus
		log.Infof("Team is #%d in solving this task", Store.passed[p]+1)
		if Store.passed[p] < len(passBonus) {
			log.Infof("Awarding %d bonus points", passBonus[Store.passed[p]])
			delta += passBonus[Store.passed[p]]
		}
		Store.passed[p]++
		Store.Teams[i].Trials[p].Passed = true

	}
	// if the problem is marked as special the reward is doubled
	if p == Store.Teams[i].Special && !(p == Store.Teams[i].Special && MainTicker.ElapsedTime().Minutes() <= 10 && !Store.Teams[i].setSpecial) {
		log.Infof("Problem was marked as special. The award is doubled")
		delta *= 2
		Store.Teams[i].SpecialScore += delta
	}

	// if the team solved all problems, give bonus
	unsolved := false
	for _, t := range Store.Teams[i].Trials {
		if !t.Passed {
			unsolved = true
			break
		}
	}
	if !unsolved {
		Store.Teams[i].finished = true;
		log.Infof("Team is #%d in solving all tasks", Store.finished+1)
		if Store.finished < len(finishBonus) {
			log.Infof("Awarding %d bonus points", finishBonus[Store.finished])
			delta += finishBonus[Store.finished]
		}
		Store.finished++
	}
	log.Infof("Final score is %d", delta)
	Store.Teams[i].Score += delta
	Store.Teams[i].Trials[p].No++
}

func fineTeam(i, s int) {
	logrus.WithFields(logrus.Fields{
		"team":   i,
		"points": s,
	}).Info("Team was fined")
	Store.Teams[i].Score -= s
}

func disqualifyTeam(i int) {
	logrus.WithField("team", i).Info("Team was disqualified")
	Store.Teams[i].Disqualified = true
}
