package biz

import (
	"context"
	"errors"
	"log"
)

var (
	ErrLevelGrade = errors.New("level Grade invalid")
)

// LevelInfo 级别等级信息
type LevelInfo struct {
	LevelGrade int // 级别等级
	// ......其它可能存在的信息
}

// NewLevelInfo 级别等级信息
func NewLevelInfo(level int) (LevelInfo, error) {
	if level < 0 {
		return LevelInfo{}, ErrLevelGrade
	}
	return LevelInfo{
		LevelGrade: level,
	}, nil
}

type LevelInfoRepo interface {
	Find(ctx context.Context, userID string) (*LevelInfo, error)
}

type LevelInfoUseCase struct {
	repo LevelInfoRepo
}

func NewLevelInfoUseCase(repo LevelInfoRepo) *LevelInfoUseCase {
	return &LevelInfoUseCase{repo: repo}
}

// GetLevelInfo 获取用户级别等级信息
func (v *LevelInfoUseCase) GetLevelInfo(ctx context.Context, userID string) (LevelInfo, error) {
	levelData, err := v.repo.Find(ctx, userID)
	if err != nil {
		log.Printf("GetLevelInfo err %v \n", err)
		if errors.Is(err, ErrUserNotFound) {
			return LevelInfo{}, ErrUserNotFound
		}
		return LevelInfo{}, ErrorBizInternal
	}
	return NewLevelInfo(levelData.LevelGrade)
}
