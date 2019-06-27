package models

import (
	"context"
	//"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type QuestionnaireModel struct {
	db *mongo.Database
}

const (
	QUESTIONNAIRE_ID_KEY string = "_id"
	TITLE_KEY            string = "title"
	QUESTION_KEY         string = "questions"
	TOPIC_KEY            string = "topic"
	ANSWER_KEY           string = "answers"
	OPTION_KEY           string = "option"
	COUNT_KEY            string = "count"
)

type Answer struct {
	Option string `bson:option`
	Count  int    `bson:count`
}

type Question struct {
	Topic   string   `bson:topic`
	Answers []Answer `bson:answers`
}

type QuestionnaireDoc struct {
	Title     string     `bson:"questionnaire_name"`
	Questions []Question `bson:questions`
}

type SimpleQuestion struct {
	Topic   string
	Options []string
}

type SimpleQuestionnaire struct {
	Title     string
	Questions []SimpleQuestion
}

// 使用/创建 collection, 初始化子 model
func NewQuestionnaireModel(db *mongo.Database) *QuestionnaireModel {
	return &QuestionnaireModel{db}
}

// 创建一个新的问卷
// 输入参数为问卷的json数据，将json数据转换成一个string，调用unmarshal来解析
func (m *QuestionnaireModel) CreateNewQuestionnaire(q *QuestionnaireDoc) (qid string) {
	id, errInsert := m.db.Collection(QuestionnaireCollectionName).InsertOne(context.TODO(), q)
	lib.AssertErr(errInsert)
	lib.Assert(id != nil, "unknown_error")
	log.Debug().Msg(fmt.Sprintf("insert a questionnaire with id = %v", id.InsertedID))
	return id.InsertedID.(primitive.ObjectID).Hex()
}

// 获得一个问卷的题目，用于填写问卷
// 输入参数为问卷的id
// 返回指定的问卷，不包括统计数据
func (m *QuestionnaireModel) GetQuestionnaire(qid string) (q *SimpleQuestionnaire) {
	objID, err := primitive.ObjectIDFromHex(qid)
	lib.AssertErr(err)
	tempQuestionnaire := &QuestionnaireDoc{}
	res := m.db.Collection(QuestionnaireCollectionName).FindOne(
		context.TODO(),
		bson.D{{
			QUESTIONNAIRE_ID_KEY,
			objID,
		}},
	)
	lib.Assert(res != nil, "no_such_questionnaire")
	lib.AssertErr(res.Decode(tempQuestionnaire))
	q = &SimpleQuestionnaire{}
	q.Title = tempQuestionnaire.Title
	for _, tempQuestion := range tempQuestionnaire.Questions {
		var allOptions []string
		for _, tempAnswer := range tempQuestion.Answers {
			allOptions = append(allOptions, tempAnswer.Option)
		}
		q.Questions = append(q.Questions, SimpleQuestion{tempQuestion.Topic, allOptions})
	}
	return
}

// 返回完整的问卷，即题目和问卷的回答统计
// 输入参数为问卷的id
// 返回指定的问卷，包括统计数据
func (m *QuestionnaireModel) GetFullQuestionnaire(qid string) (q *QuestionnaireDoc) {
	objID, err := primitive.ObjectIDFromHex(qid)
	lib.AssertErr(err)
	q = &QuestionnaireDoc{}
	res := m.db.Collection(QuestionnaireCollectionName).FindOne(
		context.TODO(),
		bson.D{{
			QUESTIONNAIRE_ID_KEY,
			objID,
		}},
	)
	lib.Assert(res != nil, "no_such_questionnaire")
	lib.Assert(res.Decode(q) == nil, "unknown_error")
	return
}

// 向问卷添加一条记录
// 输入为一个QuestionnaireDoc
// 不返回参数
func (m *QuestionnaireModel) AddOneRecord(qid string, q *QuestionnaireDoc) {
	// q := QuestionnaireDoc{}
	// errMarshal := json.Unmarshal([]byte(inputJson), q)
	// lib.AssertErr(errMarshal)
	objID, err := primitive.ObjectIDFromHex(qid)
	lib.AssertErr(err)
	stringHead := strings.Join([]string{QUESTION_KEY, ".", ANSWER_KEY}, "")
	stringOption := strings.Join([]string{stringHead, ".", OPTION_KEY}, "")
	stringCount := strings.Join([]string{stringHead, ".", COUNT_KEY}, "")
	for _, tempQuestion := range q.Questions {
		for _, tempAnswer := range tempQuestion.Answers {
			if tempAnswer.Count != 0 {
				filter := bson.D{
					{QUESTIONNAIRE_ID_KEY, objID},
					{stringOption, tempAnswer.Option},
				}
				updater := bson.D{{
					"$inc", bson.D{
						{stringCount, tempAnswer.Count},
					},
				}}
				res, err := m.db.Collection(QuestionnaireCollectionName).UpdateOne(
					context.TODO(),
					filter,
					updater,
				)
				lib.AssertErr(err)
				log.Debug().Msg(fmt.Sprintf("update result: %v", res))
			}
		}
	}
	return
}
