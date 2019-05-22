package models

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/configs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const (
	UserCollectionName       = "users"
	DelegationCollectionName = "delegations"
)

var model *Model

// Model 数据库实例
type Model struct {
	DB   *mongo.Database
	User *UserModel
}

// 连接到数据库
func InitDB(config *configs.DBConfig) error {
	model = &Model{}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return err
	}
	// test connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	log.Info().Msg("Successful connect to server")

	model.DB = client.Database(config.DBName)
	model.User = NewUserModel(model.DB)

	return nil
}

// access the model object
func GetModel() *Model {
	return model
}
