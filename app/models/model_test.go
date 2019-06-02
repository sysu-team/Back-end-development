package models

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/configs"
	"github.com/sysu-team/Back-end-development/lib"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"testing"
	"time"
)

func initDB(host, port, dbname string) {
	err := InitDB(&configs.DBConfig{
		Host:   host,
		Port:   port,
		DBName: dbname,
	})
	lib.AssertErr(err)
	log.Info().Msg("init done")
}

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

func TestRunner(t *testing.T) {
	initDB("127.0.0.1", "27017", "swsad_weapp")
	dm := GetModel().Delegation
	did := dm.CreateNewDelegation("wxm", "wxm's food", "food", 200)
	res := dm.GetDelegationPreview(1, 10)
	log.Debug().Msg(fmt.Sprintf("len %v, value %v", len(res), res))
	res1 := dm.GetSpecificDelegation(did)
	log.Debug().Msg(fmt.Sprintf("should get a delegation %v", res1))

	//for _, doc:= range res {
	//	t.Log(doc)
	// }
}

func deleleAllDoc() {
	// 需要数据库已连接上
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	lib.AssertErr(err)

	lib.AssertErr(client.Ping(context.TODO(), readpref.Primary()))

	// 把表都删除掉
	lib.AssertErr(client.Database("swsad_weapp").Collection(DelegationCollectionName).Drop(context.TODO()))
	lib.AssertErr(client.Database("swsad_weapp").Collection(UserCollectionName).Drop(context.TODO()))
}

func TestMisc(t *testing.T) {
	//res, err := resty.R().Get("https://www.gstatic.com/webp/gallery3/1.png")
	//lib.Assert(err)
}
