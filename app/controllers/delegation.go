package controllers

import (
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/services"
	"github.com/sysu-team/Back-end-development/lib"
	"strconv"
)

// UserController 用户控制
type DelegationController struct {
	BaseController
	// 使用的是 interface 而不是 struct
	Server services.DelegationService
}

// BindUserController 绑定用户控制器
func BindDelegationController(app *iris.Application) {
	delegationRoute := mvc.New(app.Party("/delegations"))

	// 使用 Register 来初始化 UserController 中的 Filed
	// 全局只有一个  sessions ，每一个连接都会生成一个 session
	delegationRoute.Register(services.NewDelegationService(), getSession().Start)
	delegationRoute.Handle(new(DelegationController))
}

func (c *DelegationController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/", "Get")
	b.Handle("GET", "/{param1:string}", "GetBy")
	b.Handle("POST", "/", "Post", withLogin)
	b.Handle("PUT", "/{param1:string}", "Put", withLogin)
}

// 返回所有活跃的委托
func (c *DelegationController) Get() {
	log.Debug().Msg(fmt.Sprintf("page : %v, limit: %v", c.Ctx.URLParam("page"), c.Ctx.URLParam("limit")))
	page, err := strconv.Atoi(c.Ctx.URLParam("page"))
	lib.AssertErr(err)
	limit, err := strconv.Atoi(c.Ctx.URLParam("limit"))
	lib.AssertErr(err)
	lib.Assert(page > 0 && limit > 0, "invalid_params")
	res := c.Server.GetDelegationPreview(page, limit)
	c.JSON(200, res, lib.Page{Page: page, Limit: limit, Total: len(res)})
}

// 获取特定的委托
func (c *DelegationController) GetBy(delegationID string) {
	// 检查参数的合法性
	lib.Assert(MatchDelegationID(delegationID), "invalid_params")
	c.JSON(200, c.Server.GetSpecificDelegation(delegationID))
}

// todo
func MatchDelegationID(delegationID string) bool {
	return true
}

// 创建委托
// 1. 检验用户是否已经登陆
// 2. 委托是否合法
// todo: testing
func (c *DelegationController) Post() {
	body := &services.DelegationInfo{}
	lib.Assert(c.Ctx.ReadJSON(body) == nil, "invalid_params")
	lib.Assert(body.Publisher == c.Session.Get(IdKey), "invalid_user")
	c.Server.CreateDelegation(body)
	c.JSON(200)
}

// 接受委托
// 1. 检验该委托是否存在
// 2. 检验委托是否已经被接受了
// 3. 检验是否满足接受的委托的条件( 具体条件待定 ）
func (c *DelegationController) Put(delegationID string) {
	c.Server.ReceiveDelegation(c.Session.GetString(IdKey), delegationID)
	c.JSON(200)
}