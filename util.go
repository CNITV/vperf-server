package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BasicError sets the status and writes the name of the status as a message
func BasicError(c *gin.Context, status int) {
	c.String(status, strconv.Itoa(status)+" "+http.StatusText(status))
}

func BodyAsNumber(c *gin.Context) (int, error) {
	b, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}

	return n, nil
}
