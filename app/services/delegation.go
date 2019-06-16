package services

import (
	"time"

	"github.com/sysu-team/Back-end-development/app/models"
	"github.com/sysu-team/Back-end-development/lib"
)

// DelegationService 用户逻辑
type DelegationService interface {
	GetDelegationPreview(page, limit, state int) []DelegationPreviewWrapper
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

type DelegationPreviewWrapper struct {
	Id          string `json:"id"`
	Name        string
	Description string
	Reward      float64
	Deadline    int64
}

func (ds *delegationService) GetDelegationPreview(page, limit, state int) []DelegationPreviewWrapper {
	tmp := ds.delegationModel.GetDelegationPreviewByState(int64(page), int64(limit), state)
	res := make([]DelegationPreviewWrapper, 0, len(tmp))
	for _, v := range tmp {
		res = append(res, DelegationPreviewWrapper{
			v.Id.Hex(),
			v.Name,
			v.Description,
			v.Reward,
			v.Deadline,
		})
	}
	return res
}

type DelegationInfoWrapper struct {
	models.DelegationDoc
	PublisherName string
	ReceiverName  string
}

func (ds *delegationService) GetSpecificDelegation(delegationID string) *DelegationInfoWrapper {
	doc := ds.delegationModel.GetSpecificDelegation(delegationID)
	receiverName := ""
	if doc.ReceiverId != "" {
		receiverName = ds.userModel.GetUserByOpenID(doc.ReceiverId).Name
	}
	return &DelegationInfoWrapper{
		*doc,
		ds.userModel.GetUserByOpenID(doc.PublisherId).Name,
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
	lib.Assert(delegation.PublisherId != receiverID, "invalid_receiver_same_as_publisher", 401)
	lib.Assert(delegation.ReceiverId == "", "invalid_delegation_already_received", 402)
	lib.Assert(delegation.Deadline < time.Now().Unix(), "invalid_delegation_timeout", 402)
	ds.delegationModel.ReceiveDelegation(delegationID, receiverID)
}

// 判断这个委托是否处于活跃状态
// 没有被接受 + 没有过期
func (ds *delegationService) isActiveDelegation(delegationID string) bool {
	delegation := ds.GetSpecificDelegation(delegationID)
	return delegation.Deadline < time.Now().Unix() && delegation.ReceiverId == ""
}

// 取消委托
func (ds *delegationService) CancelDelegation(cancelerID, delegationID string) {
	// 先检查该用户是否有资格取消该委托,必须发布者本人才能取消
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.Publisher == cancelerID, "invalid_canceler_not_cancelled_by_pulisher", 401)
	// TODO:目前似乎还没有用户的分类，如果是管理员应该也可以取消？
	// 检查该委托是否已经被取消/结束，该情况下无法取消
	lib.Assert(delegation.State != 2, "invalid_delegation_already_canceled", 402)
	lib.Assert(delegation.State != 4, "invalid_delegation_already_done", 402)
	ds.delegationModel.CancelDelegation(delegationID)
}

// 完成委托
func (ds *delegationService) FinishDelegation(finisherID, delegationID string) {
	// 首先检查该用户是否有资格完成该委托，必须接收者本人才能完成
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.Receiver == finisherID, "invalid_canceler_not_finished_by_receiver", 401)
	// TODO:发布者可以完成吗？管理员可以完成吗？
	// 检查该委托是否已经被取消/结束，该情况下无法完成
	lib.Assert(delegation.State != 2, "invalid_delegation_already_canceled", 402)
	lib.Assert(delegation.State != 4, "invalid_delegation_already_done", 402)
	ds.delegationModel.FinishDelegation(delegationID)
}
