package controllers

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
	"github.com/sysu-team/Back-end-development/app/configs"
	"github.com/sysu-team/Back-end-development/lib"
	"time"
)

type CommonRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

const (
	IdKey         = "id"
	IdTimeKey     = "idTime"
	OFFLINE_DEBUG = true
)

// singleton
var sessionManager *sessions.Sessions

type BaseController struct {
	Ctx     iris.Context
	Session *sessions.Session
}

func (c *BaseController) JSON(v ...interface{}) {
	// TODO: 应该设置成基类 controller 的方法
	// v... 后面的三个点不能省略
	// todo: 参数解包的原因
	lib.JSON(c.Ctx, v...)
}

type PageQuery struct {
	Page  *int `form:"page"`
	Limit *int `form:"limit"`
}

// InitSession 初始化 Session
func InitSession(config *configs.SessionConfig) {
	sessionManager = sessions.New(sessions.Config{
		Cookie: config.Key,
	})
}

// NewApp 创建服务器实例并绑定控制器
func NewApp() *iris.Application {
	app := iris.New()
	// 注册中间件，顺序重要
	// panic handler
	// recover from any http-relative panics
	app.Use(recover.New())
	// log the requests to the terminal.
	app.Use(logger.New())
	// error handler 错误集中处理
	app.Use(lib.NewErrorHandler())
	BindUserController(app)
	BindDelegationController(app)
	return app
}

func getSession() *sessions.Sessions {
	if sessionManager == nil {
		// 生成默认 Session
		sessionManager = sessions.New(sessions.Config{
			Cookie: "cddwxm",
		})
	}
	return sessionManager
}

// 常见中间件
// 一些接口需要微信授权状态
func withLogin(ctx iris.Context) {
	session := sessionManager.Start(ctx)
	id := session.GetString(IdKey)
	idTime := session.GetInt64Default(IdTimeKey, 0)
	lib.Assert(id != "" && idTime != 0 && time.Now().Unix()-idTime <= 86400, "invalid_token", 401)
	ctx.Values().Set(IdKey, id)
	ctx.Next()
}
