package services

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/models"
	"github.com/sysu-team/Back-end-development/lib"
)

type QuestionnaireService interface {
	GetQuestionnairePreview(delegationID string) *models.SimpleQuestionnaire
	GetFullQuestionnaire(userID, delegationID string) *models.QuestionnaireDoc
	// CreateQuestionnaire(doc *models.QuestionnaireDoc)
	AddRecord(userID, delegationID string, doc *QuestionnaireInfo)
}

func NewQuestionnaireService() QuestionnaireService {
	return &questionnaireService{
		models.GetModel().Delegation,
		models.GetModel().Questionnaire,
	}
}

type questionnaireService struct {
	delegationModel    *models.DelegationModel
	questionnaireModel *models.QuestionnaireModel
}

type QuestionnaireInfo struct {
	Title     string            `json:"Title"`
	Questions []models.Question `json:"questions"`
}

// 获得用于填写的问卷，只包含问题，不包含统计数据
// 输入的参数：委托的id
// 输出的参数：不包含统计数据的问卷
func (qs *questionnaireService) GetQuestionnairePreview(delegationID string) *models.SimpleQuestionnaire {
	delegation := qs.delegationModel.GetSpecificDelegation(delegationID)
	qid := delegation.QuestionnaireID
	log.Debug().Msg(fmt.Sprintf("不带统计的问卷: %+v", qs.questionnaireModel.GetQuestionnaire(qid)))
	return qs.questionnaireModel.GetQuestionnaire(qid)
}

// 获得完整问卷
// 输入的参数：委托的id
// 输出的参数：完整问卷
func (qs *questionnaireService) GetFullQuestionnaire(userID, delegationID string) *models.QuestionnaireDoc {
	delegation := qs.delegationModel.GetSpecificDelegation(delegationID)
	lib.Assert(delegation.PublisherID == userID, "invalid_full_questionnaire_not_get_by_publisher", 401)
	return qs.questionnaireModel.GetFullQuestionnaire(delegation.QuestionnaireID)
}

// 添加一个问卷填写的记录
// TODO:需要条件：必须为已接受的用户
// 输入参数：完整的一次问卷
// 无输出
func (qs *questionnaireService) AddRecord(userID, delegationID string, doc *QuestionnaireInfo) {
	delegation := qs.delegationModel.GetSpecificDelegation(delegationID)
	flag := 0
	for _, tempReceiverID := range delegation.ReceiverID {
		if tempReceiverID == userID {
			flag = 1
		}
	}
	lib.Assert(flag == 1, "invalid_not_add_by_current_receiver", 401)
	// log.Debug().Msg(fmt.Sprintf("填写的问卷: %+v", doc))
	oldQuestionnaire := qs.questionnaireModel.GetFullQuestionnaire(delegation.QuestionnaireID)
	for questionIndex, tempQuestion := range doc.Questions {
		for answerIndex, tempAnswer := range tempQuestion.Answers {
			oldQuestionnaire.Questions[questionIndex].Answers[answerIndex].Count += tempAnswer.Count
		}
	}

	qs.questionnaireModel.AddOneRecord(delegation.QuestionnaireID, oldQuestionnaire.Questions)
}
