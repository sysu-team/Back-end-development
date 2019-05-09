package controllers

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
	"github.com/sysu-team/Back-end-development/app/configs"
)

type CommonRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// singleton
var sessionManager *sessions.Sessions

// InitSession 初始化 Session
func InitSession(config configs.SessionConfig) {
	sessionManager = sessions.New(sessions.Config{
		Cookie: config.Key,
	})
}

// NewApp 创建服务器实例并绑定控制器
func NewApp() *iris.Application {
	app := iris.New()
	// recover from any http-relative panics
	app.Use(recover.New())
	// log the requests to the terminal.
	app.Use(logger.New())

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
