package services

import (
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/models"
)

// UserService 用户逻辑
type UserService interface {
	Register(name, studentNumber, openid string)
	HasRegistered(openid string) bool
	FindUserByOpenID(openid string) *models.UserDoc
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
	})
}

// 返回对应的用户
func (s *userService) FindUserByOpenID(openid string) *models.UserDoc {
	user := s.model.GetUserByOpenID(openid)
	log.Debug().Interface("service get user:", user)
	return user
}

// todo
func (s *userService) HasRegistered(openid string) bool {
	user := s.FindUserByOpenID(openid)
	//log.Debug().Interface("check user's register status " + openid, user)
	return user != nil
}
