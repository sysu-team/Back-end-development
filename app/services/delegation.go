package services

import (
	"github.com/sysu-team/Back-end-development/app/models"
	"github.com/sysu-team/Back-end-development/lib"
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
	model *models.DelegationModel
}

type DelegationPreviewWrapper struct {
	Id          string `json:"id"`
	Name        string
	Description string
	Reward      float64
}

func (ds *delegationService) GetDelegationPreview(page, limit int) []DelegationPreviewWrapper {
	tmp := ds.model.GetDelegationPreview(int64(page), int64(limit))
	res := make([]DelegationPreviewWrapper, 0, len(tmp))
	for _, v := range tmp {
		res = append(res, DelegationPreviewWrapper{
			v.Id.Hex(),
			v.Name,
			v.Description,
			v.Reward,
		})
	}
	return res
}

func (ds *delegationService) GetSpecificDelegation(delegationID string) *models.DelegationDoc {
	return ds.model.GetSpecificDelegation(delegationID)
}

//// info
type DelegationInfo struct {
	Publisher string `json:"publisher"`
	//StartTime time.Time `json:"start_time"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Reward      float64 `json:"reward"`
}

// todo: 基本的检查
func (ds *delegationService) CreateDelegation(info *DelegationInfo) {
	ds.model.CreateNewDelegation(info.Publisher, info.Name, info.Description, info.Reward)
}

//  接受委托
func (ds *delegationService) ReceiveDelegation(receiverID, delegationID string) {
	// 判断委托接收者是否合法的, 实际上不需要，前面已经通过 withLogin 检验过，只有合法的用户才能 login
	lib.Assert(ds.isActiveDelegation(delegationID), "invalid_delegation", 401)
	ds.model.ReceiveDelegation(delegationID, receiverID)
}

// 判断这个委托是否处于活跃状态
func (ds *delegationService) isActiveDelegation(delegationID string) bool {
	// todo
	return true
}
