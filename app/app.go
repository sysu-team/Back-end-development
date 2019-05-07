package app

import (
	"github.com/kataras/iris"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/configs"
	"github.com/sysu-team/Back-end-development/app/controllers"
	"os"
)

// Run 程序入口
func Run(configPath string) {
	// 初始化日志
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// 读取配置
	var config configs.Config
	config.GetConf(configPath)
	// 初始化各种服务
	//initService(config)
	// 启动服务器
	app := controllers.NewApp()

	//if config.Dev {
	//	app.Logger().SetLevel("debug")
	//	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	//} else {
	//	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	//}

	//if err := app.Run(iris.Addr(config.HTTP.Host + ":" + config.HTTP.Port)); err != nil {
	//	panic(err)
	//}
	if err := app.Run(iris.Addr("localhost:9999")); err != nil {
		panic(err)
	}
}
