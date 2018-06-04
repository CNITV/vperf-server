package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	configFile = "config.yaml"
	logFile    = "log.json"
)

func bailOut() {
	go func() {
		logrus.Warn("Stopping in one second")
		time.Sleep(time.Second)
		Log.Save()
		os.Exit(0)
	}()
}

func handleSigint() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		bailOut()
	}()
}

func startContest() {
	MainTicker.Start()
	logrus.Info("Contest started")
}

func reset(start bool) {
	initStorage(Conf)
	initTicker()
	initLog(logFile)
	if start {
		startContest()
	}
}

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		logrus.SetLevel(logrus.DebugLevel)
	}

	loadConfig(configFile)
	reset(false)
	handleSigint()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, false))
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AddAllowHeaders("authorization")
	r.Use(cors.New(corsConfig))
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
			BasicError(c, http.StatusServiceUnavailable)
			c.Abort()
			return
		}
		c.Next()
	})

	admin.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	admin.POST("/start", func(c *gin.Context) {
		startContest()
	})

	admin.POST("/stop", func(c *gin.Context) {
		MainTicker.Stop()
		logrus.Info("Contest stopped")
		bailOut()
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
			BasicError(c, http.StatusServiceUnavailable)
			return
		}
		setSpecial(i, p)
		Log.Push(EventSetSpecial, map[string]int{"team_id": i, "task_id": p})
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
			BasicError(c, http.StatusServiceUnavailable)
			return
		}
		ans, err := BodyAsNumber(c)
		if err != nil {
			BasicError(c, http.StatusBadRequest)
			return
		}
		submitAnswer(i, p, ans)
		Log.Push(EventSubmitAnswer, map[string]int{"team_id": i, "task_id": p, "answer": ans})
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
		fineTeam(i, s)
		Log.Push(EventFineTeam, map[string]int{"team_id": i, "points": s})
	})

	admin.DELETE("/team/:id", func(c *gin.Context) {
		i, err := strconv.Atoi(c.Param("id"))
		if err != nil || i < 0 || i >= len(Store.Teams) {
			BasicError(c, http.StatusBadRequest)
			return
		}
		rawForget := c.Query("forget")
		forget, err := strconv.ParseBool(rawForget)
		if err != nil {
			BasicError(c, http.StatusBadRequest)
			return
		}

		disqualifyTeam(i)
		Log.Push(EventDisqualifyTeam, map[string]int{"team_id": i})

		if forget {
			toDelete := []int{}
			for ei, e := range Log.Entries {
				val, ok := e.Params["team_id"]
				if ok && val == i && e.Event != EventDisqualifyTeam {
					toDelete = append(toDelete, ei)
				}
			}
			Log.Delete(toDelete...)
			Log.Save()
			reset(true)
		}
	})

	admin.GET("/log", func(c *gin.Context) {
		c.JSON(http.StatusOK, Log.Entries)
	})

	admin.DELETE("/log/:id", func(c *gin.Context) {
		i, err := strconv.Atoi(c.Param("id"))
		if err != nil || i < 0 || i >= len(Log.Entries) {
			BasicError(c, http.StatusBadRequest)
			return
		}
		Log.Delete(i)
		Log.Save()
		reset(true)
	})

	r.Run(":1031")
}
