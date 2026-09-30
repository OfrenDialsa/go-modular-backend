package config

import "github/OfrenDialsa/go-gin-starter/docs"

func InitSwagger(env *EnvironmentVariable) {
	docs.SwaggerInfo.Title = env.Swagger.Title
	docs.SwaggerInfo.Description = env.Swagger.Description
	docs.SwaggerInfo.Version = env.Swagger.Version
	docs.SwaggerInfo.Host = env.Swagger.Host
	docs.SwaggerInfo.BasePath = "/docs"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}
