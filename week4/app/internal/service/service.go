package service

import (
	"ceshi/week4/app/internal/biz"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewUserService)

type UserService struct {
	uc *biz.UserBaseUseCase
	vc *biz.LevelInfoUseCase
}

func NewUserService(uc *biz.UserBaseUseCase, vc *biz.LevelInfoUseCase) *UserService {
	return &UserService{
		uc: uc,
		vc: vc,
	}
}
