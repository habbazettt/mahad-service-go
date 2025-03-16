package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

func InitLogger() {
	logFormat := os.Getenv("LOG_FORMAT")

	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}
