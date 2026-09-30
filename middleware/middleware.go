package middleware

import (
	"github/OfrenDialsa/go-gin-starter/lib"
	"time"

	"github.com/gin-gonic/gin"
)

type Middleware interface {
	Validate(roles ...lib.Role) gin.HandlerFunc
	EmailVerified() gin.HandlerFunc
	RateLimit(limit int, window time.Duration) gin.HandlerFunc
	Prometheus() gin.HandlerFunc
}
