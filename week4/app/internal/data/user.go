package data

import (
	"ceshi/week4/app/internal/biz"
	"context"
)

type userRepo struct {
	// 这里暂时不进行实际的数据库操作和映射
}

func NewUserRepo() biz.UserBaseRepo {
	return &userRepo{}
}

func (r *userRepo) Find(ctx context.Context, userID string) (*biz.UserBase, error) {
	// TODO 数据库操作
	return &biz.UserBase{Name: "liwei", Age: 27}, nil
}
