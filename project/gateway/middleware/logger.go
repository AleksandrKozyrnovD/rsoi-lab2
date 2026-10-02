package middleware

import (
	"github.com/gin-gonic/gin"
	"log/slog"
)

func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		errorMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()

		slog.Info("request",
			"method", method,
			"path", path,
			"raw_query", raw,
			"status", status,
			"client_ip", clientIP,
			"user_agent", userAgent,
			"errors", errorMsg,
		)
	}
}

