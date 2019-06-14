package services

import (
	"github.com/sysu-team/Back-end-development/app/models"
	"github.com/sysu-team/Back-end-development/lib"
	"time"
)

// DelegationService 用户逻辑
type DelegationService interface {
	GetDelegationPreview(page, limit int) []DelegationPreviewWrapper
	GetSpecificDelegation(delegationID string) *models.DelegationDoc
	CreateDelegation(info *DelegationInfo)
	ReceiveDelegation(receiverID, delegationID string)
}

func NewDelegationService() DelegationService {
	return &delegationService{
		models.GetModel().Delegation,
	}
}

type delegationService struct {
	delegationModel *models.DelegationModel
}

type DelegationPreviewWrapper struct {
	Id          string `json:"id"`
	Name        string
	Description string
	Reward      float64
	Deadline    int64
}

func (ds *delegationService) GetDelegationPreview(page, limit int) []DelegationPreviewWrapper {
	tmp := ds.delegationModel.GetDelegationPreview(int64(page), int64(limit))
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

func (ds *delegationService) GetSpecificDelegation(delegationID string) *models.DelegationDoc {
	return ds.delegationModel.GetSpecificDelegation(delegationID)
}

//// info
type DelegationInfo struct {
	Publisher string `json:"publisher"`
	//StartTime time.Time `json:"start_time"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Reward      float64 `json:"reward"`
	Deadline    int64   `json:"deadline"`
	Type        string  `json:"type"`
}

// todo: 基本的检查
func (ds *delegationService) CreateDelegation(info *DelegationInfo) {
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
	lib.Assert(delegation.Publisher != receiverID, "invalid_receiver_same_as_publisher", 401)
	lib.Assert(delegation.Receiver == "", "invalid_delegation_already_received", 402)
	lib.Assert(delegation.Deadline < time.Now().Unix(), "invalid_delegation_timeout", 402)
	ds.delegationModel.ReceiveDelegation(delegationID, receiverID)
}

// 判断这个委托是否处于活跃状态
// 没有被接受 + 没有过期
func (ds *delegationService) isActiveDelegation(delegationID string) bool {
	delegation := ds.GetSpecificDelegation(delegationID)
	return delegation.Deadline < time.Now().Unix() && delegation.Receiver == ""
}
