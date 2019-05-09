package controllers

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"
	"github.com/rs/zerolog/log"
	"gopkg.in/resty.v1"
)

// UserController 用户控制
type UserController struct {
	Ctx     iris.Context
	Session *sessions.Session
}

// BindUserController 绑定用户控制器
func BindUserController(app *iris.Application) {
	userRoute := mvc.New(app.Party("/users"))
	// 初始化全局使用的 session
	// TODO: session 使用的地方
	userRoute.Register(getSession().Start)
	userRoute.Handle(new(UserController))
}

type LoginReq struct {
	Code string `json:"code"`
}

type WxSessionRes struct {
	OpenId     string `json:"code"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// 从微信后端获取对应的 openid 和 session key 等数据
// TODO iris - MVC 架构如何处理返回值 ?
func (c *UserController) PostUsersSession() (CommonRes, int) {
	// read the the params from request

	log.Debug().Str("info", "invoke post login")

	// 获取请求中的code
	body := LoginReq{}
	if err := c.Ctx.ReadJSON(&body); err != nil {
		return CommonRes{Code: 400, Msg: "invalid_params"}, 400
	}
	log.Debug().Str("code", body.Code)
	// 使用 code 和 appid 到微信后台获取对应的
	// GET https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=JSCODE&grant_type=authorization_code
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"appid":      "wx296231dd4ca34b67",
			"secret":     "e6dd8a2438d2d8201e3b6e8da0fa34b0",
			"js_code":    body.Code,
			"grant_type": "authorization_code",
		}).
		SetResult(&WxSessionRes{}).
		Get("https://api.weixin.qq.com/sns/jscode2session")
	if resp == nil || err == nil {
		log.Error().Err(err)
		return CommonRes{Code: 401, Msg: "error occur when get "}, 400
	}
	log.Debug().Msg(resp.String())
	wx_res := resp.Result().(*WxSessionRes)
	c.Ctx.Header("session_id", wx_res.SessionKey)
	// 维护自定义登陆状态
	c.Session.Set("session_key", wx_res.SessionKey)
	c.Session.Set("open_id", wx_res.OpenId)

	return CommonRes{Code: 200, Msg: "ok"}, 200

}
