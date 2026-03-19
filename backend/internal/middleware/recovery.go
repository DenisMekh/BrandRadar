package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("panic recovered: %v | method: %s | path: %s",
					err, c.Request.Method, c.Request.URL.Path)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"data": nil,
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "internal server error",
					},
				})
			}
		}()

		c.Next()
	}
}
