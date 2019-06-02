package app

import (
	"github.com/json-iterator/go/extra"
	"github.com/kataras/iris"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/configs"
	"github.com/sysu-team/Back-end-development/app/controllers"
	"github.com/sysu-team/Back-end-development/app/models"
	"os"
)

type LoginReq struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Code     string `json:"code"`
	Captcha  string `json:"captcha"`
}

func initService(config configs.Config) {
	controllers.InitSession(&config.HTTP.Session)
}

// Run 程序入口
func Run(configPath string) {
	// 初始化日志, 添加输出行号
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})
	// 读取配置
	var config configs.Config
	config.GetConf(configPath)
	// 初始化各种服务
	// 初始化 session
	controllers.InitSession(&config.HTTP.Session)

	// 初始化 Json 设置
	// 自动转换成小写下划线风格
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)

	// 初始化 database
	if err := models.InitDB(&config.Db); err != nil {
		panic(err)
	}

	//OFFLINE_DEBUG = config.Offline
	// 启动服务器
	app := controllers.NewApp()

	if config.Dev {
		app.Logger().SetLevel("debug")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
	log.Debug().Msg(config.HTTP.Host)
	if err := app.Run(iris.Addr(config.HTTP.Host + ":" + config.HTTP.Port)); err != nil {
		panic(err)
	}
}
