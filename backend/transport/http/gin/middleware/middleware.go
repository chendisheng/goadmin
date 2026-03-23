package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := coremiddleware.NormalizeRequestID(c.GetHeader(coremiddleware.RequestIDHeader))
		if requestID == "" {
			requestID = coremiddleware.GenerateRequestID()
		}
		c.Set(coremiddleware.RequestIDContextKey, requestID)
		c.Header(coremiddleware.RequestIDHeader, requestID)
		c.Next()
	}
}

func Recovery(log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID := requestIDFromGinContext(c)
				status, body := response.Failure(fmt.Errorf("panic: %v", recovered), requestID)
				log.Error("panic recovered",
					zap.Any("panic", recovered),
					zap.String("request_id", requestID),
					zap.ByteString("stack", debug.Stack()),
				)
				c.AbortWithStatusJSON(status, body)
			}
		}()
		c.Next()
	}
}

func AccessLog(log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		requestID := requestIDFromGinContext(c)
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.String("raw_path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
			log.Error("http request", fields...)
			return
		}
		log.Info("http request", fields...)
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Request-ID",
			"X-Tenant-ID",
		}, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		}, ", "))
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func requestIDFromGinContext(c *gin.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if requestID, ok := value.(string); ok && requestID != "" {
			return requestID
		}
	}
	return ""
}
