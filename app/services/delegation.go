package services

import (
	//"fmt"
	"time"

	//"github.com/rs/zerolog/log"
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
		models.GetModel().Questionnaire,
	}
}

type delegationService struct {
	delegationModel    *models.DelegationModel
	userModel          *models.UserModel
	questionnaireModel *models.QuestionnaireModel
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
	if len(doc.ReceiverID) != 0 {
		receiverName = ds.userModel.GetUserByOpenID(doc.ReceiverID[0]).Name
	}
	return &DelegationInfoWrapper{
		*doc,
		ds.userModel.GetUserByOpenID(doc.PublisherID).Name,
		receiverName,
	}
}

//// info
type DelegationInfoReq struct {
	Publisher     string                   `json:"publisher"`
	Name          string                   `json:"name"`
	Description   string                   `json:"description"`
	Reward        int                      `json:"reward"`
	Deadline      int64                    `json:"deadline"`
	Type          string                   `json:"type"`
	MaxNumber     int                      `json:"max_number"`
	Questionnaire *models.QuestionnaireDoc `json:"questionnaire"`
}

// todo: 基本的检查
func (ds *delegationService) CreateDelegation(info *DelegationInfoReq) {
	// 检查积分是否满足要求
	publisher := ds.userModel.GetUserByOpenID(info.Publisher)
	newCredit := publisher.Credit - info.MaxNumber*info.Reward
	lib.Assert(newCredit >= 0, "no_enough_credit_to_create_delegation", 401)
	// log.Debug().Msg(fmt.Sprintf("questionnaire: %+v", info.Questionnaire))
	// log.Debug().Msg(fmt.Sprintf("answers: %+v", info.Questionnaire.Questions[0].Answers))
	var qid string
	if info.Type == "填写问卷" {
		qid = ds.questionnaireModel.CreateNewQuestionnaire(info.Questionnaire)
	}
	ds.delegationModel.CreateNewDelegation(
		info.Publisher,
		info.Name,
		info.Description,
		info.Reward,
		info.Deadline,
		info.Type,
		qid,
		info.MaxNumber,
	)
	ds.userModel.SetCreditByOpenID(info.Publisher, newCredit)
}

//  接受委托
func (ds *delegationService) ReceiveDelegation(receiverID, delegationID string) {
	// 判断委托接收者是否合法的, 委托和接收者不能是同一个人
	delegation := ds.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.PublisherID != receiverID, "invalid_receiver_same_as_publisher", 401)
	for _, tempReceiverID := range delegation.ReceiverID {
		lib.Assert(tempReceiverID != receiverID, "invalid_delegation_already_receive", 402)
	}
	lib.Assert(delegation.DelegationState == 0, "invalid_delegation_already_received", 402)
	lib.Assert(delegation.Deadline > time.Now().Unix(), "invalid_delegation_timeout", 403)
	// 计算是否有足够的积分进行接受时的预冻结，不够则报错
	receiver := ds.userModel.GetUserByOpenID(receiverID)
	newCredit := receiver.Credit - delegation.Reward
	lib.Assert(newCredit >= 0, "not_enough_credit_to_receive", 403)
	var newState uint8
	if delegation.CurrentNumber == delegation.MaxNumber-1 {
		newState = 1
	}
	ds.delegationModel.ReceiveDelegation(delegationID, receiverID, newState)
	ds.userModel.SetCreditByOpenID(receiverID, newCredit)
}

// 判断这个委托是否处于活跃状态
// 没有被接受 + 没有过期
func (ds *delegationService) isActiveDelegation(delegationID string) bool {
	delegation := ds.GetSpecificDelegation(delegationID)
	return delegation.Deadline < time.Now().Unix() && delegation.CurrentNumber < delegation.MaxNumber
}

// 取消委托
func (ds *delegationService) CancelDelegation(cancelerID, delegationID string) {
	// 先检查该用户是否有资格取消该委托
	// 对于委托的发布者，可以取消
	// 对于委托的接受者，可以放弃
	delegation := ds.GetSpecificDelegation(delegationID)
	flag := 0
	for _, tempReceiverID := range delegation.ReceiverID {
		if tempReceiverID == cancelerID {
			flag = 1
		}
	}
	lib.Assert(delegation.PublisherID == cancelerID || flag == 1, "invalid_canceler_not_cancelled_by_pulisher_or_receiver")
	// 检查该委托是否能被取消
	lib.Assert(delegation.DelegationState == 0 || delegation.DelegationState == 1, "invalid_delegation_state_cannot_be_canceled", 402)
	publisher := ds.userModel.GetUserByOpenID(delegation.PublisherID)
	// 还没有被接受，预冻结的积分返还发布者
	var newState uint8 = 2
	if delegation.CurrentNumber == 0 {
		ds.userModel.SetCreditByOpenID(publisher.OpenID, publisher.Credit+delegation.Reward)
		ds.delegationModel.SetDelegationState(delegationID, newState)
	} else {
		// 已接受后，取消方损失所有的预冻结积分，被取消方获得双方预冻结的所有积分
		if delegation.PublisherID == cancelerID {
			for _, tempReceiverID := range delegation.ReceiverID {
				receiver := ds.userModel.GetUserByOpenID(tempReceiverID)
				ds.userModel.SetCreditByOpenID(tempReceiverID, receiver.Credit+2*delegation.Reward)
				ds.delegationModel.DeleteReceiver(delegationID, tempReceiverID, newState)
			}
		} else {
			ds.userModel.SetCreditByOpenID(publisher.OpenID, publisher.Credit+2*delegation.Reward)
			if delegation.MaxNumber != 1 {
				newState = 1
			}
			ds.delegationModel.DeleteReceiver(delegationID, cancelerID, newState)
		}
	}

	// TODO:判断委托是否已经过DDL
}

var timer *time.Timer

// 完成委托
func (ds *delegationService) FinishDelegation(finisherID, delegationID string) {
	// 首先检查该用户是否有资格完成该委托，必须接收者本人才能完成
	delegation := ds.GetSpecificDelegation(delegationID)
	flag := 0
	for _, tempReceiver := range delegation.ReceiverID {
		if finisherID == tempReceiver {
			flag = 1
		}
	}
	lib.Assert(flag == 1 || delegation.PublisherID == finisherID, "invalid_canceler_not_finished_by_receiver", 401)
	FinishByPublisher := func() {
		var newState uint8 = 4
		rewardCoe := 2
		if delegation.DelegationState == 1 {
			rewardCoe = 1
			if delegation.MaxNumber != 1 {
				newState = 1
			}
		}
		for _, tempReceiverID := range delegation.ReceiverID {
			receiver := ds.userModel.GetUserByOpenID(tempReceiverID)
			ds.userModel.SetCreditByOpenID(tempReceiverID, receiver.Credit+rewardCoe*delegation.Reward)
		}
		ds.delegationModel.SetDelegationState(delegationID, newState)
	}
	// 对于不同的用户，检查委托的状态的不同条件
	if delegation.PublisherID == finisherID {
		// 当发布者确认完成后，将双方预冻结的积分给接受者
		lib.Assert(delegation.DelegationState == 3, "invalid_delegation_not_pending", 402)
		FinishByPublisher()
	} else {
		// 接受者完成，等待发布者确认
		lib.Assert(delegation.DelegationState == 1, "invalid_delegation_not_accepted", 402)
		// 若所有的用户都完成了
		if delegation.MaxNumber == 1 {
			ds.delegationModel.SetDelegationState(delegationID, 3)
			timer = time.AfterFunc(3600*time.Second, FinishByPublisher)
		} else {
			// 若不是，则删掉一个用户
			ds.delegationModel.DeleteReceiver(delegationID, finisherID, 1)
		}
	}
	//log.Debug().Msg("End finish")
	// TODO:判断委托是否已经过DDL
	// ds.delegationModel.SetDelegationState(delegationID, 4)
}
