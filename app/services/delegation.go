package services

import (
	"time"

	"github.com/sysu-team/Back-end-development/app/models"
	"github.com/sysu-team/Back-end-development/lib"
)

// DelegationService 用户逻辑
type DelegationService interface {
	GetDelegationPreview(page, limit, state int) []models.DelegationPreviewWrapper
	GetSpecificDelegation(delegationID string) *DelegationInfoWrapper
	CreateDelegation(info *DelegationInfoReq)
	ReceiveDelegation(receiverID, delegationID string)
	CancelDelegation(cancelerID, delegationID string)
	FinishDelegation(finisherID, delegationID string)
}

func NewDelegationService() DelegationService {
	return &delegationService{
		models.GetModel().Delegation,
		models.GetModel().User,
	}
}

type delegationService struct {
	delegationModel *models.DelegationModel
	userModel       *models.UserModel
}

func (ds *delegationService) GetDelegationPreview(page, limit, state int) []models.DelegationPreviewWrapper {
	return ds.delegationModel.GetDelegationPreviewByState(int64(page), int64(limit), state)
}

type DelegationInfoWrapper struct {
	models.DelegationDoc
	PublisherName string
	ReceiverName  string
}

func (ds *delegationService) GetSpecificDelegation(delegationID string) *DelegationInfoWrapper {
	doc := ds.delegationModel.GetSpecificDelegation(delegationID)
	receiverName := ""
	if doc.ReceiverID != "" {
		receiverName = ds.userModel.GetUserByOpenID(doc.ReceiverID).Name
	}
	return &DelegationInfoWrapper{
		*doc,
		ds.userModel.GetUserByOpenID(doc.PublisherID).Name,
		receiverName,
	}
}

//// info
type DelegationInfoReq struct {
	Publisher string `json:"publisher"`
	//StartTime time.Time `json:"start_time"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Reward      float64 `json:"reward"`
	Deadline    int64   `json:"deadline"`
	Type        string  `json:"type"`
}

// todo: 基本的检查
func (ds *delegationService) CreateDelegation(info *DelegationInfoReq) {
	ds.delegationModel.CreateNewDelegation(
		info.Publisher,
		info.Name,
		info.Description,
		info.Reward,
		info.Deadline,
		info.Type)
}

//  接受委托
func (ds *delegationService) ReceiveDelegation(receiverID, delegationID string) {
	// 判断委托接收者是否合法的, 委托和接收者不能是同一个人
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.PublisherID != receiverID, "invalid_receiver_same_as_publisher", 401)
	lib.Assert(delegation.ReceiverID == "", "invalid_delegation_already_received", 402)
	lib.Assert(delegation.Deadline > time.Now().Unix(), "invalid_delegation_timeout", 403)
	ds.delegationModel.ReceiveDelegation(delegationID, receiverID)
}

// 判断这个委托是否处于活跃状态
// 没有被接受 + 没有过期
func (ds *delegationService) isActiveDelegation(delegationID string) bool {
	delegation := ds.GetSpecificDelegation(delegationID)
	return delegation.Deadline < time.Now().Unix() && delegation.ReceiverID == ""
}

// 取消委托
func (ds *delegationService) CancelDelegation(cancelerID, delegationID string) {
	// 先检查该用户是否有资格取消该委托
	// 对于委托的发布者，可以取消
	// 对于委托的接受者，可以放弃
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.PublisherID == cancelerID || delegation.ReceiverID == cancelerID, "invalid_canceler_not_cancelled_by_pulisher_or_receiver", 401)
	// TODO:目前似乎还没有用户的分类，如果是管理员应该也可以取消？
	// TODO:判断委托是否已经过DDL
	// 检查该委托是否已经被取消/结束，该情况下无法取消
	lib.Assert(delegation.DelegationState != 2, "invalid_delegation_already_canceled", 402)
	lib.Assert(delegation.DelegationState != 4, "invalid_delegation_already_done", 402)
	ds.delegationModel.SetDelegationState(delegationID, 2)
}

// 完成委托
func (ds *delegationService) FinishDelegation(finisherID, delegationID string) {
	// 首先检查该用户是否有资格完成该委托，必须接收者本人才能完成
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.PublisherID == finisherID || delegation.ReceiverID == finisherID, "invalid_canceler_not_finished_by_publisher_or_receiver", 401)
	// 对于不同的用户，检查委托的状态的不同条件
	if delegation.PublisherID == finisherID {
		lib.Assert(delegation.DelegationState == 3, "invalid_delegation_not_pending", 402)
		ds.delegationModel.SetDelegationState(delegationID, 4)
	} else {
		lib.Assert(delegation.DelegationState == 1, "invalid_delegation_not_accepted", 402)
		ds.delegationModel.SetDelegationState(delegationID, 3)
	}
	// TODO:判断委托是否已经过DDL
	// ds.delegationModel.SetDelegationState(delegationID, 4)
}
