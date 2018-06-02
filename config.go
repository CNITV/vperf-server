package main

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var Conf *Config

type Config struct {
	AdminPass string `yaml:"admin_pass"`
	Teams     []struct {
		Name string `yaml:"name"`
	} `yaml:"teams"`

	// Time is in minutes
	Time int `yaml:"time"`

	Solutions           []int `yaml:"solutions"`
	DefaultTeamScore    int   `yaml:"default_team_score"`
	DefaultProblemScore int   `yaml:"default_problem_score"`
}

func loadConfig(path string) {
	logrus.WithField("path", path).Info("Reading config file")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.WithError(err).Fatal("Failed reading config file")
	}

	err = yaml.Unmarshal(b, &Conf)
	if err != nil {
		logrus.WithError(err).Fatal("Failed parsing config file")
	}

	logrus.Debugf("%#v", Conf)
}
