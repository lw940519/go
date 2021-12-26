package data

import (
	"ceshi/week4/app/internal/biz"
	"context"
)

type levelInfoRepo struct {
	// 这里暂时不进行实际的数据库操作和映射
}

func NewLevelInfoRepo() biz.LevelInfoRepo {
	return &levelInfoRepo{}
}

func (r *levelInfoRepo) Find(ctx context.Context, userID string) (*biz.LevelInfo, error) {
	// TODO 数据库操作
	return &biz.LevelInfo{LevelGrade: 1}, nil
}
