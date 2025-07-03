package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const TraceIDKey = "traceId"

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := c.GetHeader("X-Trace-Id")
		if traceId == "" {
			traceId = uuid.New().String()
		}
		c.Set(TraceIDKey, traceId)
		c.Writer.Header().Set("X-Trace-Id", traceId)
		c.Next()
	}
}
