package router

import (
	"fmt"
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/docs"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupSwagger(base *gin.RouterGroup, env *config.EnvironmentVariable) {
	docs.SwaggerInfo.Title = env.Swagger.Title
	docs.SwaggerInfo.Description = env.Swagger.Description
	docs.SwaggerInfo.Version = env.Swagger.Version
	docs.SwaggerInfo.Host = env.Swagger.Host
	docs.SwaggerInfo.BasePath = env.Swagger.BasePath

	uiPath := "/swagger"
	fullPath := fmt.Sprintf("http://%s%s/index.html", env.Swagger.Host, uiPath)
	base.GET(uiPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Info().
		Str("host", env.Swagger.Host).
		Str("ui_url", fullPath).
		Msg("Swagger documentation initialized")
}
