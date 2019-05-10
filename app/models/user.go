package models

import "go.mongodb.org/mongo-driver/mongo"

type UserModel struct {
	db *mongo.Database
}

type UserDoc struct {
	OpenID        string
	Name          string
	StudentNumber string
}

// 使用/创建 collcetion
// TODO 是否需要 ？
func NewCollection() {
	// create new collection
}

func (m *UserModel) AddUser(newUser UserDoc) {
	// insert user into doc

}
