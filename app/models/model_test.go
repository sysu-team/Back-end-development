package models

import (
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
	test := GetModel().User
	t.Log(test)

	res := test.AddUser(&UserDoc{
		"abc",
		"wxm",
		"110",
	})

	if err != nil {
		t.Error(err)
	}
	t.Log(res)

	user, err := test.GetUserByName("wxm")

	if err != nil {
		t.Fatal(err)
	}

	// t.Log(user)

	if user == nil {
		t.Log("no such user")
	} else if user.Name != "wxm" {
		t.Fatal("some thing went wrong")
	}

}

func Test(t *testing.T) {
}
