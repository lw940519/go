package service

import (
	"ceshi/week2/dao"
	"database/sql"
	"github.com/pkg/errors"
)

func GetUserList() ([]dao.User, error) {
	var userModel dao.User
	list, err := userModel.GetUserList()
	if errors.Cause(err) == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, errors.Wrap(err, "bar failed")
	}
	return list, nil
}
