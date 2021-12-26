package service

import (
	"ceshi/week4/api"
	"ceshi/week4/app/internal/biz"
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
)

var (
	ErrUserServiceInternal = errors.New("user Service internal")
)

func (u *UserService) UserMy(ctx context.Context, r *api.UserRequest) (*api.UserMyReply, error) {
	// 领域业务拼接
	userBase := biz.UserBase{}
	levelInfo := biz.LevelInfo{}

	wg, sctp := errgroup.WithContext(ctx)

	wg.Go(func() error {
		userB, err := u.uc.GetUserBase(sctp, r.UserID)
		if err != nil {
			return err
		}
		userBase = userB
		return nil
	})

	wg.Go(func() error {
		levelData, err := u.vc.GetLevelInfo(sctp, r.UserID)
		if err != nil {
			return err
		}
		levelInfo = levelData
		return nil
	})

	if err := wg.Wait(); err != nil {
		if errors.Is(err, biz.ErrUserNotFound) {
			return &api.UserMyReply{}, err
		} else {
			return &api.UserMyReply{}, ErrUserServiceInternal
		}
	} else {
		return &api.UserMyReply{Name: userBase.Name, Level: int64(levelInfo.LevelGrade)}, nil
	}
}
