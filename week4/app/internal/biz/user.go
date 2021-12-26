package biz

import (
	"context"
	"errors"
	"log"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserNameInvalid = errors.New("name invalid")
	ErrUserAgeInvalid  = errors.New("age invalid")
)

// UserBase 用户基本信息
type UserBase struct {
	Name string // 姓名
	Age  int32  // 年龄
}

func NewUser(name string, age int32) (UserBase, error) {
	if len(name) <= 0 {
		return UserBase{}, ErrUserNameInvalid
	}
	if age < 0 {
		return UserBase{}, ErrUserAgeInvalid
	}

	return UserBase{
		Name: name,
		Age:  age,
	}, nil
}

type UserBaseRepo interface {
	Find(ctx context.Context, userID string) (*UserBase, error)
}
type UserBaseUseCase struct {
	repo UserBaseRepo
}

func NewUserBaseUseCase(repo UserBaseRepo) *UserBaseUseCase {
	return &UserBaseUseCase{repo: repo}
}

// GetUserBase 获取用户基本信息
func (u *UserBaseUseCase) GetUserBase(ctx context.Context, userID string) (UserBase, error) {
	user, err := u.repo.Find(ctx, userID)
	if err != nil {
		log.Printf("GetUserBase err %v \n", err)
		if errors.Is(err, ErrUserNotFound) {
			return UserBase{}, ErrUserNotFound
		}
		return UserBase{}, ErrorBizInternal
	}
	return NewUser(user.Name, user.Age)
}
