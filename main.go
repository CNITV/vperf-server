package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	configFile = "config.yaml"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		logrus.SetLevel(logrus.DebugLevel)
	}

	loadConfig(configFile)
	initStorage(Conf)
	initTicker()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, false))
	r.Use(cors.Default())
	r.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	})

	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/authors", func(c *gin.Context) {
		c.String(http.StatusOK, "Tudor Roman & Ciprian Ionescu")
	})

	r.GET("/status", func(c *gin.Context) {
		Store.Time = int(MainTicker.RemainingTime().Seconds())
		Store.Running = MainTicker.Running
		b, err := json.Marshal(Store)
		if err != nil {
			logrus.Panic(err)
		}
		c.JSON(http.StatusOK, json.RawMessage(b))
	})

	admin := r.Group("/admin", gin.BasicAuth(gin.Accounts{
		"admin": Conf.AdminPass,
	}))
	tAdmin := admin.Group("/", func(c *gin.Context) {
		if !MainTicker.Running {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		c.Next()
	})

	admin.POST("/start", func(c *gin.Context) {
		MainTicker.Start()
		logrus.Info("Contest started")
	})

	admin.POST("/pause", func(c *gin.Context) {
		Store.PauseReason = c.Query("reason")
		MainTicker.Stop()
		logrus.WithField("reason", Store.PauseReason).Info("Contest paused")
	})

	admin.POST("/stop", func(c *gin.Context) {
		MainTicker.Stop()
		logrus.Info("Contest stopped")
	})

	tAdmin.PUT("/team/:id/special", func(c *gin.Context) {
		i, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			BasicError(c, http.StatusNotFound)
			return
		}
		p, err := BodyAsNumber(c)
		if err != nil {
			BasicError(c, http.StatusBadRequest)
			return
		}
		if i < 0 || i >= len(Store.Teams) || p < 0 || p >= len(Store.Problems) {
			BasicError(c, http.StatusBadRequest)
			return
		}

		if MainTicker.ElapsedTime().Minutes() >= float64(Conf.GraceTime) {
			logrus.WithFields(logrus.Fields{
				"team": i,
				"task": p,
			}).Info("Ignored special task set request. Time's out")
			c.Status(http.StatusServiceUnavailable)
			return
		}
		Store.Teams[i].Special = p
	})

	tAdmin.POST("/team/:id/submit/:problem_no", func(c *gin.Context) {
		i, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			BasicError(c, http.StatusNotFound)
			return
		}
		p, err := strconv.Atoi(c.Param("problem_no"))
		if err != nil {
			BasicError(c, http.StatusNotFound)
			return
		}
		if i < 0 || i >= len(Store.Teams) || p < 0 || p >= len(Store.Problems) {
			BasicError(c, http.StatusBadRequest)
			return
		}

		if MainTicker.ElapsedTime().Minutes() < float64(Conf.GraceTime) {
			logrus.WithFields(logrus.Fields{
				"team": i,
				"task": p,
			}).Info("Ignored submit answer request. Not the time.")
			c.Status(http.StatusServiceUnavailable)
			return
		}
		ans, err := BodyAsNumber(c)
		logrus.Debug(ans, err)
		if err != nil {
			BasicError(c, http.StatusBadRequest)
			return
		}
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
		if p == Store.Teams[i].Special {
			log.Infof("Problem was marked as special. The award is doubled")
			delta *= 2
			Store.Teams[i].SpecialScore += delta
		}
		log.Infof("Final score is %d", delta)
		Store.Teams[i].Score += delta
		Store.Teams[i].Trials[p].No++
	})

	tAdmin.POST("/team/:id/fine", func(c *gin.Context) {
		i, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			BasicError(c, http.StatusNotFound)
			return
		}
		if i < 0 || i >= len(Store.Teams) {
			BasicError(c, http.StatusBadRequest)
			return
		}
		s, err := BodyAsNumber(c)
		if err != nil {
			BasicError(c, http.StatusBadRequest)
			return
		}
		Store.Teams[i].Score -= s
	})

	r.Run(":1031")
}
