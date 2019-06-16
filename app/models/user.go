package models

import (
	"context"
	"github.com/sysu-team/Back-end-development/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserModel struct {
	db *mongo.Database
}

// 所有字段名字都是小写的
type UserDoc struct {
	OpenID        string `bson:"open_id"`
	Name          string `bson:"name"`
	StudentNumber string `bson:"student_num"`
	Credit        int    `bson:"credit"`
}

// 使用/创建 collcetion, 初始化子 model
func NewUserModel(db *mongo.Database) *UserModel {
	// create new collection
	return &UserModel{db}
}

func (m *UserModel) AddUser(newUser *UserDoc) string {
	// insert user doc into
	res, err := m.db.Collection(UserCollectionName).InsertOne(context.TODO(), newUser)
	lib.AssertErr(err)
	return res.InsertedID.(primitive.ObjectID).String()
}

// 返回nil，代表没有找到对应的用户
func (m *UserModel) GetUserByName(name string) *UserDoc {
	return m.findUserBy("name", name)
}

// 返回nil代表没有找到该 openid 对应的用户
func (m *UserModel) GetUserByOpenID(openid string) *UserDoc {
	return m.findUserBy("open_id", openid)
}

// 返回nil代表没有找到该 openid 对应的用户
func (m *UserModel) GetUserByStudentNum(studentNum string) *UserDoc {
	return m.findUserBy("student_num", studentNum)
}

func (m *UserModel) findUserBy(key, value string) *UserDoc {
	filter := bson.D{{key, value}}
	res := &UserDoc{}
	// 找不到对应的用户抛出 no document error
	err := m.db.Collection(UserCollectionName).FindOne(context.TODO(), filter).Decode(res)
	if err == mongo.ErrNoDocuments {
		return nil
	}
	lib.AssertErr(err)
	return res
}
