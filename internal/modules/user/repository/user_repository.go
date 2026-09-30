package repository

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, filter dto.GetUserFilter) (*model.User, error)
	Update(ctx context.Context, userId string, fields dto.UpdateUserFilter) error
	IsExist(ctx context.Context, filter dto.GetUserFilter) (bool, error)
	Delete(ctx context.Context, userId string) error
}
