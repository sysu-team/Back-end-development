package controllers

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
	"github.com/sysu-team/Back-end-development/app/configs"
	"github.com/sysu-team/Back-end-development/lib"
)

type CommonRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

const (
	IdKey         = "id"
	OFFLINE_DEBUG = true
)

// singleton
var sessionManager *sessions.Sessions

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
// 需要是已经微信授权的用户才能进行注册
func withLogin(ctx iris.Context) {
	//session := sessionManager.Start(ctx)
	id := ctx.GetHeader(IdKey)
	if id == "" {
		ctx.StatusCode(401)
		_, _ = ctx.JSON(CommonRes{Code: 401, Msg: "invalid_token"})
		return
	}
	ctx.Values().Set(IdKey, id)
	ctx.Next()
}
