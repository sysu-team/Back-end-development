package services

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/models"
	"github.com/sysu-team/Back-end-development/lib"
)

// UserService 用户逻辑
type UserService interface {
	Register(name, studentNumber, openid string)
	HasRegistered(openid string) bool
	FindUserByOpenID(openid string) *models.UserDoc
	GetUserInfo(openid string) *UserInfo
}

func NewUserService() UserService {
	return &userService{
		models.GetModel().User,
	}
}

type userService struct {
	model *models.UserModel
}

func (s *userService) Register(name, studentNumber, openid string) {
	s.model.AddUser(&models.UserDoc{
		OpenID:        openid,
		Name:          name,
		StudentNumber: studentNumber,
		Credit:        0,
	})
}

// 返回对应的用户
func (s *userService) FindUserByOpenID(openid string) *models.UserDoc {
	user := s.model.GetUserByOpenID(openid)
	log.Debug().Msg(fmt.Sprintf("service get user by open id  %v: %v", openid, user))
	return user
}

// 按照学号来搜索用户 / 寻找是否有这个学号的用户
func (s *userService) FindUserByStudentNum(studentNum string) *models.UserDoc {
	user := s.model.GetUserByStudentNum(studentNum)
	log.Debug().Msg(fmt.Sprintf("service get user by student number %v: %v", studentNum, user))
	return user
}

// login result
type UserInfo struct {
	Name          string `json:"name"`
	StudentNumber string `json:"studentNumber"`
	Credit        int    `json:"credit"`
}

//
func (s *userService) GetUserInfo(openid string) *UserInfo {
	user := s.FindUserByOpenID(openid)
	lib.Assert(user != nil, "unknown_error")
	return &UserInfo{
		user.Name,
		user.StudentNumber,
		user.Credit,
	}
}

// 确认是否注册
func (s *userService) HasRegistered(openid string) bool {
	user := s.FindUserByOpenID(openid)
	//log.Debug().Interface("check user's register status " + openid, user)
	return user != nil
}
