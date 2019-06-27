package models

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DelegationModel struct {
	db *mongo.Database
}

type EnumDelegationState uint8

const (
	Published EnumDelegationState = 0
	Accepted  EnumDelegationState = 1
	Canceled  EnumDelegationState = 2
	Pending   EnumDelegationState = 3
	Finished  EnumDelegationState = 4
	ANY       EnumDelegationState = 0xff
)

const (
	RECEIVER_ID_KEY       string = "receiver_id"
	PUBLISHER_ID_KEY      string = "publisher_id"
	DELETAION_ID_KEY      string = "_id"
	DELEGATAION_STATE_KEY string = "delegation_state"
	CURRENT_NUMBER_KEY    string = "current_number"
	MAX_NUMBER_KEY        string = "max_number"
)

// 所有字段名字都是小写 + 下划线连接
type DelegationDoc struct {
	PublisherID     string              `bson:"publisher_id"`
	ReceiverID      []string            `bson:"receiver_id"`
	DelegationName  string              `bson:"delegation_name"`
	StartTime       int64               `bson:"start_time"`
	DelegationState EnumDelegationState `bson:"delegation_state"`
	Reward          int                 `bson:"reward"`
	Description     string              `bson:"description"`
	Deadline        int64               `bson:"deadline"`
	DelegationType  string              `bson:"delegation_type"`
	QuestionnaireID string              `bson:"questionnaire_id"`
	MaxNumber       int                 `bson:"max_number"`
	CurrentNumber   int                 `bson:"current_number"`
}

type delegationPreviewDoc struct {
	Name        string `json:"delegation_name" bson:"delegation_name"`
	Description string
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Reward      int
	Deadline    int64
}

// 使用/创建 collection, 初始化子 model
func NewDelegationModel(db *mongo.Database) *DelegationModel {
	// create new collection
	return &DelegationModel{db}
}

// 创建新的委托
// 状态未活跃的委托没有接收者
// 返回委托 did
func (m *DelegationModel) CreateNewDelegation(publisher, name, description string, reward int, deadline int64, delegationType string, qid string, max int) (did string) {
	var receivers = make([]string, 0, max)
	id, err := m.db.Collection(DelegationCollectionName).InsertOne(context.TODO(), DelegationDoc{
		publisher,
		receivers,
		name,
		time.Now().Unix(),
		Published,
		reward,
		description,
		deadline,
		delegationType,
		qid,
		max,
		0,
	})
	lib.AssertErr(err)
	lib.Assert(id != nil, "unknown_error")
	log.Debug().Msg(fmt.Sprintf("insert a doc with id = %v", id.InsertedID))
	return id.InsertedID.(primitive.ObjectID).Hex()
}

type DelegationPreviewWrapper struct {
	Id          string `json:"id"`
	Name        string
	Description string
	Reward      int
	Deadline    int64
}

// with key and value
type DelegationFilters = bson.D

// 获取委托预览
// 按照分页的规格返回特定的委托
// 长度为0代表没有找到 不会返回 error，只有一个数据来源，error 的处理直接在中间件中处理
func (m *DelegationModel) GetDelegationPreviewByState(page, limit int64, state int) []DelegationPreviewWrapper {
	return m.getDelegationPreviewListBy(page, limit, DelegationFilters{
		{DELEGATAION_STATE_KEY, state},
	})
}

// 获取用户接受的委托的处于某个状态的委托
// ANY 意味着对状态没有要求
// 状态的检查应该再 service 层中完成
func (m *DelegationModel) GetUserAcceptedDelegationPreviewWithState(page, limit int64, userID string, state EnumDelegationState) []DelegationPreviewWrapper {
	if state != ANY {
		return m.getDelegationPreviewListBy(page, limit, DelegationFilters{
			{RECEIVER_ID_KEY, userID},
			{DELEGATAION_STATE_KEY, state},
		})
	}
	return m.getDelegationPreviewListBy(page, limit, DelegationFilters{
		{RECEIVER_ID_KEY, userID},
	})
}

// 获取用户发布的委托
func (m *DelegationModel) GetUserPublishDelegationPreviewWithState(page, limit int64, userID string, state EnumDelegationState) []DelegationPreviewWrapper {
	if state != ANY {
		return m.getDelegationPreviewListBy(page, limit, DelegationFilters{
			{PUBLISHER_ID_KEY, userID},
			{DELEGATAION_STATE_KEY, state},
		})
	}
	return m.getDelegationPreviewListBy(page, limit, DelegationFilters{
		{PUBLISHER_ID_KEY, userID},
	})
}

// 获取用户完成的委托
// 分成两部分
// 1. 用户发布的 -》 已经完成整个流程了
// 2. 用户接受的 -》 等待整个流程
func (m *DelegationModel) GetUserPendingDelegationPreviewWithState(page, limit int64, userID string, state EnumDelegationState) []DelegationPreviewWrapper {
	return m.getDelegationPreviewListBy(page, limit, DelegationFilters{
		{RECEIVER_ID_KEY, userID},
		{DELEGATAION_STATE_KEY, state},
	})
}

// 接受委托
// 输入object id, 和接受委托人
// 更新数据库中的委托信息
// 可能抛出的错误：
// 1. 这是一个已经被接受的委托
// 2. 不存在该委托
func (m *DelegationModel) ReceiveDelegation(delegationID string, receiverID string, state uint8) {
	objID, err := primitive.ObjectIDFromHex(delegationID)
	lib.AssertErr(err)
	res, err := m.db.Collection(DelegationCollectionName).UpdateOne(
		context.TODO(),
		bson.D{{
			DELETAION_ID_KEY,
			objID,
		}},
		bson.D{
			{
				"$addToSet", bson.D{
					{RECEIVER_ID_KEY, receiverID},
				},
			},
			{
				"$set", bson.D{
					{DELEGATAION_STATE_KEY, state},
				},
			},
			{
				"$inc", bson.D{
					{CURRENT_NUMBER_KEY, 1},
				},
			},
		},
	)
	lib.AssertErr(err)
	log.Debug().Msg(fmt.Sprintf("update result: %v", res))
	return
}

func (m *DelegationModel) SetDelegationState(delegationID string, state uint8) {
	objID, err := primitive.ObjectIDFromHex(delegationID)
	lib.AssertErr(err)
	res, err := m.db.Collection(DelegationCollectionName).UpdateOne(
		context.TODO(),
		bson.D{{
			DELETAION_ID_KEY,
			objID,
		}},
		bson.D{{
			"$set", bson.D{
				{DELEGATAION_STATE_KEY, state},
			},
		}},
	)
	lib.AssertErr(err)
	log.Debug().Msg(fmt.Sprintf("set result: %v", res))
	return
}

// 获取委托详细情况
// 根据委托 id 获取委托
// Object ID 获取和返回
func (m *DelegationModel) GetSpecificDelegation(uniqueID string) (d *DelegationDoc) {
	objID, err := primitive.ObjectIDFromHex(uniqueID)
	lib.AssertErr(err)
	d = &DelegationDoc{}
	res := m.db.Collection(DelegationCollectionName).FindOne(
		context.TODO(),
		bson.D{{
			DELETAION_ID_KEY,
			objID,
		}},
	)
	lib.Assert(res != nil, "no_such_delegation")
	lib.Assert(res.Decode(d) == nil, "unknown_error")
	return
}

func (m *DelegationModel) getDelegationPreviewListBy(page, limit int64, filters DelegationFilters) []DelegationPreviewWrapper {
	res := make([]DelegationPreviewWrapper, 0, limit)
	offset := (page - 1) * limit
	cursor, err := m.db.Collection(DelegationCollectionName).
		Find(
			context.TODO(),
			filters,
			&options.FindOptions{
				Limit: &limit,
				Skip:  &offset,
			})
	if err == mongo.ErrNilDocument {
		return res
	}
	lib.AssertErr(err)
	lib.Assert(cursor != nil, "unknown_error", 401)
	defer func() {
		lib.AssertErr(cursor.Close(context.TODO()))
	}()
	for cursor.Next(context.TODO()) {
		tmp := delegationPreviewDoc{}
		// 这是一个应该直接抛出的错误
		lib.AssertErr(cursor.Decode(&tmp))
		res = append(res, DelegationPreviewWrapper{
			tmp.ID.Hex(),
			tmp.Name,
			tmp.Description,
			tmp.Reward,
			tmp.Deadline,
		})
	}
	return res
}

// 问卷的接受者取消/完成，将最大人数和当前人数各减一，同时将取消/完成者从列表中删除
func (m *DelegationModel) DeleteReceiver(delegationID, userID string, newState uint8) {
	objID, err := primitive.ObjectIDFromHex(delegationID)
	lib.AssertErr(err)
	res, err := m.db.Collection(DelegationCollectionName).UpdateOne(
		context.TODO(),
		bson.D{{
			DELETAION_ID_KEY,
			objID,
		}},
		bson.D{
			{
				"$pull", bson.D{
					{RECEIVER_ID_KEY, userID},
				},
			},
			{
				"$set", bson.D{
					{DELEGATAION_STATE_KEY, newState},
				},
			},
			{
				"$inc", bson.D{
					{CURRENT_NUMBER_KEY, -1},
					{MAX_NUMBER_KEY, -1},
				},
			},
		},
	)
	lib.AssertErr(err)
	log.Debug().Msg(fmt.Sprintf("DeleteReceiver result: %v", res))
}
