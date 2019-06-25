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
	Name        string `json:"name"`
	Description string `json:"description"`
	Reward      int    `json:"reward"`
	Deadline    int64  `json:"deadline"`
	Type        string `json:"type"`
}

// todo: 基本的检查
func (ds *delegationService) CreateDelegation(info *DelegationInfoReq) {
	// 检查积分是否满足要求
	publisher := ds.userModel.GetUserByOpenID(info.Publisher)
	newCredit := publisher.Credit - info.Reward
	lib.Assert(newCredit >= 0, "no_enough_credit_to_create_delegation", 401)
	ds.delegationModel.CreateNewDelegation(
		info.Publisher,
		info.Name,
		info.Description,
		info.Reward,
		info.Deadline,
		info.Type)
	ds.userModel.SetCreditByOpenID(info.Publisher, newCredit)
}

//  接受委托
func (ds *delegationService) ReceiveDelegation(receiverID, delegationID string) {
	// 判断委托接收者是否合法的, 委托和接收者不能是同一个人
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.PublisherID != receiverID, "invalid_receiver_same_as_publisher", 401)
	lib.Assert(delegation.ReceiverID == "" && delegation.DelegationState == 0, "invalid_delegation_already_received", 402)
	lib.Assert(delegation.Deadline > time.Now().Unix(), "invalid_delegation_timeout", 403)
	// 计算是否有足够的积分进行接受时的预冻结，不够则报错
	receiver := ds.userModel.GetUserByOpenID(receiverID)
	newCredit := receiver.Credit - delegation.Reward
	lib.Assert(newCredit >= 0, "not_enough_credit_to_receive", 403)
	ds.delegationModel.ReceiveDelegation(delegationID, receiverID)
	ds.userModel.SetCreditByOpenID(receiverID, newCredit)
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
	// 检查该委托是否能被取消
	lib.Assert(delegation.DelegationState == 0 || delegation.DelegationState == 1, "invalid_delegation_state_cannot_be_canceled", 402)
	publisher := ds.userModel.GetUserByOpenID(delegation.PublisherID)
	receiver := ds.userModel.GetUserByOpenID(delegation.ReceiverID)
	// 还没有被接受，预冻结的积分返还发布者
	if delegation.DelegationState == 0 {
		ds.userModel.SetCreditByOpenID(publisher.OpenID, publisher.Credit+delegation.Reward)
	} else {
		// 已接受后，取消方损失所有的预冻结积分，被取消方获得双方预冻结的所有积分
		if delegation.ReceiverID == cancelerID {
			ds.userModel.SetCreditByOpenID(publisher.OpenID, publisher.Credit+2*delegation.Reward)
		} else {
			ds.userModel.SetCreditByOpenID(receiver.OpenID, receiver.Credit+2*delegation.Reward)
		}
	}

	// TODO:判断委托是否已经过DDL
	ds.delegationModel.SetDelegationState(delegationID, 2)
}

var timer *time.Timer

// 完成委托
func (ds *delegationService) FinishDelegation(finisherID, delegationID string) {
	// 首先检查该用户是否有资格完成该委托，必须接收者本人才能完成
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.PublisherID == finisherID || delegation.ReceiverID == finisherID, "invalid_canceler_not_finished_by_publisher_or_receiver", 401)
	FinishByPublisher := func() {
		receiver := ds.userModel.GetUserByOpenID(delegation.ReceiverID)
		if delegation.DelegationState == 3 {
			ds.userModel.SetCreditByOpenID(receiver.OpenID, receiver.Credit+2*delegation.Reward)
		} else {
			ds.userModel.SetCreditByOpenID(receiver.OpenID, receiver.Credit+delegation.Reward)
		}
		ds.delegationModel.SetDelegationState(delegationID, 4)
	}
	// 对于不同的用户，检查委托的状态的不同条件
	if delegation.PublisherID == finisherID {
		// 当发布者确认完成后，将双方预冻结的积分给接受者
		lib.Assert(delegation.DelegationState == 3, "invalid_delegation_not_pending", 402)
		FinishByPublisher()
	} else {
		// 接受者完成，等待发布者确认
		lib.Assert(delegation.DelegationState == 1, "invalid_delegation_not_accepted", 402)
		ds.delegationModel.SetDelegationState(delegationID, 3)
		timer = time.AfterFunc(20*time.Second, FinishByPublisher)
	}
	//log.Debug().Msg("End finish")
	// TODO:判断委托是否已经过DDL
	// ds.delegationModel.SetDelegationState(delegationID, 4)
}
