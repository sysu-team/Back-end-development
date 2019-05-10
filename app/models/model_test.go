package models

import (
	"context"
	"github.com/sysu-team/Back-end-development/app/configs"
	"testing"
)

func TestInitDB(t *testing.T) {
	t.Log("test start")
	err := InitDB(&configs.DBConfig{
		Host:   "127.0.0.1",
		Port:   "27017",
		DBName: "swsad_weapp",
	})
	t.Log("after init")
	if err != nil {
		t.Error(err)
	}
	test := GetModel().DB.Collection("user")
	t.Log(test)
	res, err := test.InsertOne(context.TODO(), &UserDoc{"test", "ttt", "qq"})
	if err != nil {
		t.Error(err)
	}
	t.Log(res.InsertedID)
}
