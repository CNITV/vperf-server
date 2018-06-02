package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	configFile = "config.yaml"
)

// BasicError sets the status and writes the name of the status as a message
func BasicError(c *gin.Context, status int) {
	c.Status(status)
	c.Writer.Write([]byte(strconv.Itoa(status) + " " + http.StatusText(status)))
}

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

	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/authors", func(c *gin.Context) {
		c.String(http.StatusOK, "Tudor Roman & Ciprian Ionescu")
	})

	r.GET("/status", func(c *gin.Context) {
		Store.Time = Conf.Time*60 - int((MainTicker.Prev + MainTicker.ElapsedSinceStart()).Seconds())
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

	admin.POST("/start", func(c *gin.Context) {
		MainTicker.Start()
	})

	admin.POST("/pause", func(c *gin.Context) {
		Store.PauseReason = c.Query("reason")
		MainTicker.Stop()
	})

	admin.POST("/stop", func(c *gin.Context) {
		MainTicker.Stop()
	})

	admin.PUT("/team/:id/special", func(c *gin.Context) {
		i, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			BasicError(c, http.StatusNotFound)
			return
		}
		b, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			BasicError(c, http.StatusBadRequest)
			return
		}
		p, err := strconv.Atoi(string(b))
		if err != nil {
			BasicError(c, http.StatusBadRequest)
			return
		}
		if i < 0 || i >= len(Store.Teams) || p < 0 || p >= len(Store.Problems) {
			BasicError(c, http.StatusBadRequest)
			return
		}
		Store.Teams[i].Special = p
	})

	r.Run(":1031")
}
