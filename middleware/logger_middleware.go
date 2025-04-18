package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		stop := time.Now()

		logrus.WithFields(logrus.Fields{
			"status":  c.Response().StatusCode(),
			"method":  c.Method(),
			"path":    c.Path(),
			"latency": stop.Sub(start).String(),
		}).Info("HTTP request")

		return err
	}
}
