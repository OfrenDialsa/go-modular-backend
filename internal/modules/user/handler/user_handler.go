package handler

import "github.com/gin-gonic/gin"

type UserHandler interface {
	GetMe(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	ChangePassword(ctx *gin.Context)
	DeleteAccount(ctx *gin.Context)
	DeleteAvatar(ctx *gin.Context)
}
