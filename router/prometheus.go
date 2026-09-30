package router

import (
	"github/OfrenDialsa/go-gin-starter/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupPrometheus(env *config.EnvironmentVariable, router *gin.Engine) {
	metricsH := gin.WrapH(promhttp.Handler())

	if env.App.Mode == "prod" {
		auth := gin.BasicAuth(gin.Accounts{
			env.Prometheus.Username: env.Prometheus.Password,
		})
		router.GET("/metrics", auth, metricsH)
		return
	}

	router.GET("/metrics", metricsH)
}
