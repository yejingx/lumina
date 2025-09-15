package dao

import (
	"lumina/internal/model"
	"time"
)

type UserSpec struct {
	Id          int    `json:"id"`
	Username    string `json:"username" binding:"required"`
	Nickname    string `json:"nickname" binding:"required"`
	IsAdmin     bool   `json:"isAdmin"`
	CreatedTime string `json:"createdTime" binding:"required,datetime=RFC3339"`
}

type LoginRequest struct {
	// 用户名
	Username string `json:"username" binding:"required"`
	// 密码
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	// 登录凭证
	Token string   `json:"token"`
	User  UserSpec `json:"user"`
}

type ListUsersRequest struct {
	// 分页开始位置
	Start int `form:"start"`
	// 分页大小
	Limit int `form:"limit"`
}

type ListUsersResponse struct {
	// 用户总数
	Total int `json:"total"`
	// 用户列表
	Items []UserSpec `json:"items"`
}

type CreateUserRequest struct {
	// 用户名
	Username string `json:"username" binding:"required"`
	// 密码
	Password string `json:"password" binding:"required"`
	// 昵称
	Nickname string `json:"nickname" binding:"required"`
	// 是否管理员
	IsAdmin bool `json:"isAdmin"`
	// 部门id
	DepartmentId int `json:"departmentId" binding:"required"`
}

type CreateUserResponse struct {
	// 用户ID
	Id int `json:"id"`
}

func ToUserSpec(u *model.User) (*UserSpec, error) {
	return &UserSpec{
		Id:          u.Id,
		Username:    u.Username,
		Nickname:    u.Nickname,
		IsAdmin:     u.IsAdmin,
		CreatedTime: u.CreatedTime.Format(time.RFC3339),
	}, nil
}

func (r CreateUserRequest) ToUserModel() (*model.User, error) {
	return &model.User{
		Username: r.Username,
		Password: r.Password,
		Nickname: r.Nickname,
		IsAdmin:  r.IsAdmin,
	}, nil
}
