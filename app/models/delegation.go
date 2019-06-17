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
)

const (
	RECEIVER_ID_KEY       string = "receiver_id"
	PUBLISHER_ID_KEY      string = "publisher_id"
	DELETAION_ID_KEY      string = "_id"
	DELEGATAION_STATE_KEY string = "delegation_state"
)

// 所有字段名字都是小写 + 下划线连接
type DelegationDoc struct {
	PublisherID     string              `bson:"publisher_id"`
	ReceiverID      string              `bson:"receiver_id"`
	DelegationName  string              `bson:"delegation_name"`
	StartTime       int64               `bson:"start_time"`
	DelegationState EnumDelegationState `bson:"delegation_state"`
	Reward          float64             `bson:"reward"`
	Description     string              `bson:"description"`
	Deadline        int64               `bson:"deadline"`
	DelegationType  string              `bson:"delegation_type"`
}

type DelegationPreviewDoc struct {
	Name        string `json:"delegation_name" bson:"delegation_name"`
	Description string
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Reward      float64
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
func (m *DelegationModel) CreateNewDelegation(publisher, name, description string, reward float64, deadline int64, delegationType string) (did string) {
	id, err := m.db.Collection(DelegationCollectionName).InsertOne(context.TODO(), DelegationDoc{
		publisher,
		"",
		name,
		time.Now().Unix(),
		Published,
		reward,
		description,
		deadline,
		delegationType,
	})
	lib.AssertErr(err)
	lib.Assert(id != nil, "unknown_error")
	log.Debug().Msg(fmt.Sprintf("insert a doc with id = %v", id.InsertedID))
	return id.InsertedID.(primitive.ObjectID).Hex()
}

// 获取委托预览
// 按照分页的规格返回特定的委托
// 长度为0代表没有找到 不会返回 error，只有一个数据来源，error 的处理直接在中间件中处理
func (m *DelegationModel) GetDelegationPreviewByState(page, limit int64, state int) []DelegationPreviewDoc {
	res := make([]DelegationPreviewDoc, 0, limit)
	//findOption = options.Find()
	offset := (page - 1) * limit
	cursor, err := m.db.Collection(DelegationCollectionName).
		Find(
			context.TODO(),
			bson.D{
				{DELEGATAION_STATE_KEY, state},
			},
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
		tmp := DelegationPreviewDoc{}
		// 这是一个应该直接抛出的错误
		lib.AssertErr(cursor.Decode(&tmp))
		res = append(res, tmp)
	}
	return res
}

// 接受委托
// 输入object id, 和接受委托人
// 更新数据库中的委托信息
//
// 可能抛出的错误：
// 1. 这是一个已经被接受的委托
// 2. 不存在该委托
func (m *DelegationModel) ReceiveDelegation(delegationID string, receiverID string) {
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
				{RECEIVER_ID_KEY, receiverID},
				{DELEGATAION_STATE_KEY, Accepted},
			},
		}},
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
			"_id",
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
