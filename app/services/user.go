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
	// 获取用户相关的委托
	GetUserPendingDelegation(page, limit int, receiverUserID string) []models.DelegationPreviewWrapper
	GetUserPublishDelegation(page, limit int, publisherUserID string) []models.DelegationPreviewWrapper
	GetUserReceiveDelegation(page, limit int, receiverUserID string) []models.DelegationPreviewWrapper
}

func NewUserService() UserService {
	return &userService{
		models.GetModel().User,
		models.GetModel().Delegation,
	}
}

type userService struct {
	userModel       *models.UserModel
	delegationModel *models.DelegationModel
}

func (s *userService) Register(name, studentNumber, openid string) {
	s.userModel.AddUser(&models.UserDoc{
		OpenID:        openid,
		Name:          name,
		StudentNumber: studentNumber,
		Credit:        0,
	})
}

// 返回对应的用户
func (s *userService) FindUserByOpenID(openid string) *models.UserDoc {
	user := s.userModel.GetUserByOpenID(openid)
	log.Debug().Msg(fmt.Sprintf("service get user by open id  %v: %v", openid, user))
	return user
}

// 按照学号来搜索用户 / 寻找是否有这个学号的用户
func (s *userService) FindUserByStudentNum(studentNum string) *models.UserDoc {
	user := s.userModel.GetUserByStudentNum(studentNum)
	log.Debug().Msg(fmt.Sprintf("service get user by student number %v: %v", studentNum, user))
	return user
}

// login result
type UserInfo struct {
	Name          string `json:"name"`
	StudentNumber string `json:"studentNumber"`
	Credit        int    `json:"credit"`
}

// 获取用户信息
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

// 返回用户已完成的等待发布者确定的委托
func (s *userService) GetUserPendingDelegation(page, limit int, receiverUserID string) []models.DelegationPreviewWrapper {
	return s.delegationModel.GetUserPendingDelegationPreviewWithState(int64(page), int64(limit), receiverUserID, models.Pending)
}

// 返回用户发布的委托
func (s *userService) GetUserPublishDelegation(page, limit int, publisherUserID string) []models.DelegationPreviewWrapper {
	return s.delegationModel.GetUserPublishDelegationPreviewWithState(int64(page), int64(limit), publisherUserID, models.Published)
}

// 返回处于接受状态的还没有完成的委托
func (s *userService) GetUserReceiveDelegation(page, limit int, receiverUserID string) []models.DelegationPreviewWrapper {
	return s.delegationModel.GetUserAcceptedDelegationPreviewWithState(int64(page), int64(limit), receiverUserID, models.Accepted)
}
