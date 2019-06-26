package controllers

import (
	//"fmt"
	//"strconv"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	//"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/services"
	"github.com/sysu-team/Back-end-development/lib"
)

// 问卷控制
type QuestionnaireController struct {
	BaseController
	Server services.QuestionnaireService
}

// 绑定问卷控制器
func BindQuestionnaireController(app *iris.Application) {
	questionnaireRoute := mvc.New(app.Party("/questionnnaire"))

	questionnaireRoute.Register(services.NewQuestionnaireService(), getSession().Start)
	questionnaireRoute.Handle(new(QuestionnaireController))
}

func (c *QuestionnaireController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("PUT", "/{param1:string}", "Put", withLogin)
	b.Handle("GET", "/{param1:string}", "Get")
	b.Handle("GET", "/{param1:string}/result", "GetResult", withLogin)
}

// 填写问卷函数
// 1. 检查是否已经填写过
// 2. 检查是否接受了该问卷
func (c *QuestionnaireController) Put(delegationID string) {
	lib.Assert(c.Session.GetString(IdKey) != "", "unknown_err")
	questionnaire := &services.QuestionnaireInfo{}
	lib.Assert(c.Ctx.ReadJSON(questionnaire) == nil, "invalid_params")
	c.Server.AddRecord(c.Session.GetString(IdKey), delegationID, questionnaire)
	c.JSON(200)
}

// 获得问卷的题目，用于填写
func (c *QuestionnaireController) Get(delegationID string) {
	c.Server.GetQuestionnairePreview(delegationID)
	c.JSON(200)
}

// 获得问卷以及统计信息
// 1. 检查用户是否发布者
func (c *QuestionnaireController) GetResult(delegationID string) {
	lib.Assert(c.Session.GetString(IdKey) != "", "unknown_err")
	c.Server.GetFullQuestionnaire(c.Session.GetString(IdKey), delegationID)
	c.JSON(200)
}
