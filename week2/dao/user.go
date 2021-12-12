package dao

import "database/sql"

type User struct {
	ID int
	UserName string
}

func (u *User)GetUserList() ([]User,error) {
	var userList []User
	//链接sql数据库
	//执行sql语句
	//绑定数据
	return userList,sql.ErrNoRows
}
