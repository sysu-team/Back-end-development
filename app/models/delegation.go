package models

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type DelegationModel struct {
	db *mongo.Database
}

type DelegationState uint8

const (
	Active  DelegationState = 0
	Pending DelegationState = 1
	Done    DelegationState = 2
)

// 所有字段名字都是小写的
// TODO: marshal and unmarshal
type DelegationDoc struct {
	Publisher string
	Receiver  string // open id
	Name      string
	StartTime time.Time
	EndTime   time.Time
	State     DelegationState
}

// 使用/创建 collection, 初始化子 model
func NewDelegationModel(db *mongo.Database) *DelegationModel {
	// create new collection
	return &DelegationModel{db}
}

// 创建新的委托
func (m *DelegationModel) CreateNewDelegation(db *mongo.Database) {
	// todo
}
