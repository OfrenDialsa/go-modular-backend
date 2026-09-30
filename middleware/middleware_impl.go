package middleware

import (
	"fmt"
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/internal/infra/metrics"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/service"
	"github/OfrenDialsa/go-gin-starter/lib"
	"github/OfrenDialsa/go-gin-starter/utils"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/gin-gonic/gin"
)

type MiddlewareImpl struct {
	env        *config.EnvironmentVariable
	sessionSvc service.SessionService
	store      *cache.Cache
	mu         sync.Mutex
}

func NewMiddleware(
	env *config.EnvironmentVariable,
	sessionSvc service.SessionService,
	store *cache.Cache,
) Middleware {
	return &MiddlewareImpl{
		env:        env,
		sessionSvc: sessionSvc,
		store:      store,
	}
}

func (m *MiddlewareImpl) Validate(roles ...lib.Role) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			lib.RespondError(c, lib.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			lib.RespondError(c, lib.ErrInvalidAuthorizationFormat)
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := lib.ValidateToken(tokenString, m.env.JWT.SecretKey.Access)
		if err != nil {
			lib.RespondError(c, lib.ErrInvalidAuthorizationFormat)
			c.Abort()
			return
		}

		session, err := m.sessionSvc.GetBySessionId(c.Request.Context(), claims.SessionId)
		if err != nil || session == nil {
			lib.RespondError(c, lib.ErrInvalidAuthorizationFormat)
			c.Abort()
			return
		}

		if len(roles) > 0 {
			if !utils.ContainsRole(lib.Role(session.Role), roles) {
				lib.RespondError(c, lib.ErrForbidden)
				c.Abort()
				return
			}
		}

		c.Set("is_verified", session.EmailVerifiedAt != nil)
		c.Set("user", claims)

		c.Next()
	}
}

func (m *MiddlewareImpl) EmailVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		isVerified, exists := c.Get("is_verified")
		if !exists || isVerified == false {
			lib.RespondError(c, lib.ErrEmailNotVerified)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (m *MiddlewareImpl) RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		i := c.ClientIP()
		if user, exists := c.Get("user"); exists {
			if claims, ok := user.(*lib.JWTClaims); ok {
				i = claims.UserId
			}
		}

		key := "rate_limit:" + c.FullPath() + ":" + i

		now := time.Now()
		windowSec := window.Seconds()
		currStart := now.Truncate(window).Unix()

		var w *utils.CounterWindow

		m.mu.Lock()
		if item, found := m.store.Get(key); found {
			w = item.(*utils.CounterWindow)
		} else {
			w = &utils.CounterWindow{
				CurrWindow: currStart,
			}
			m.store.Set(key, w, window*2)
		}

		m.mu.Unlock()
		w.Mu.Lock()
		defer w.Mu.Unlock()

		if currStart > w.CurrWindow {
			if currStart-w.CurrWindow == int64(window.Seconds()) {
				w.PrevCount = w.CurrCount
			} else {
				w.PrevCount = 0
			}

			w.LastWindow = w.CurrWindow
			w.CurrWindow = currStart
			w.CurrCount = 0
		}

		timePassed := now.Sub(time.Unix(w.CurrWindow, 0)).Seconds()
		weight := (windowSec - timePassed) / windowSec
		count := int(float64(w.PrevCount)*weight) + w.CurrCount

		if count >= limit {
			nextWindowStart := w.CurrWindow + int64(window.Seconds())
			remainingSeconds := nextWindowStart - now.Unix()

			if remainingSeconds <= 0 {
				remainingSeconds = 1
			}

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("Retry-After", fmt.Sprintf("%d", remainingSeconds))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Too many requests, please try again later.",
				"error": gin.H{
					"code":        "TOO_MANY_REQUESTS",
					"retry_after": remainingSeconds,
				},
			})
			c.Abort()
			return
		}

		w.CurrCount++
		m.store.Set(key, w, window*2)

		c.Next()
	}
}

func (m *MiddlewareImpl) Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		endpoint := c.FullPath()
		if endpoint == "/metrics" || strings.HasPrefix(endpoint, "/swagger") {
			c.Next()
			return
		}

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		metrics.HTTPRequests.WithLabelValues(method, endpoint, status).Inc()
		metrics.HTTPDuration.WithLabelValues(method, endpoint).
			Observe(time.Since(start).Seconds())
	}
}
