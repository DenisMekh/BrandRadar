package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if len(c.Errors) > 0 {
			logrus.Errorf("request error [%d] %s %s %s (%s) - %v",
				status, c.Request.Method, path, query, c.ClientIP(), c.Errors.String())
			return
		}

		switch {
		case status >= 500:
			logrus.Errorf("server error [%d] %s %s %s (%s)",
				status, c.Request.Method, path, query, c.ClientIP())
		case status >= 400:
			logrus.Warnf("client error [%d] %s %s %s (%s)",
				status, c.Request.Method, path, query, c.ClientIP())
		default:
			logrus.Infof("request [%d] %s %s %s (%s) completed in %v",
				status, c.Request.Method, path, query, c.ClientIP(), latency.Round(time.Millisecond))
		}
	}
}
