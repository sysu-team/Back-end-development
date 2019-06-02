package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/rs/zerolog/log"
	"github.com/sysu-team/Back-end-development/app/services"
	"github.com/sysu-team/Back-end-development/lib"
	"gopkg.in/resty.v1"
	"time"
)

// UserController 用户控制
type UserController struct {
	BaseController
	// 使用的是 interface 而不是 struct
	Server services.UserService
}

// BindUserController 绑定用户控制器
func BindUserController(app *iris.Application) {
	userRoute := mvc.New(app.Party("/users"))

	// 使用 Register 来初始化 UserController 中的 Filed
	// 全局只有一个  sessions ，每一个连接都会生成一个 session
	userRoute.Register(services.NewUserService(), getSession().Start)
	userRoute.Handle(new(UserController))
}

func (c *UserController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("POST", "/", "Post")
	b.Handle("POST", "/session", "PostSession")
	b.Handle("DELETE", "/session", "DelSession", withLogin)
}

type LoginReq struct {
	Code string `json:"code"`
}

type WxSessionRes struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrCode    int64  `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func wxAuth(code string) *WxSessionRes {
	res := &WxSessionRes{}
	if OFFLINE_DEBUG {
		// 如果不是使用小程序后端进行实验，采用body中的code作为id
		res.OpenId = code
		return res
	}
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"appid":      "wx296231dd4ca34b67",
			"secret":     "e6dd8a2438d2d8201e3b6e8da0fa34b0",
			"js_code":    code,
			"grant_type": "authorization_code",
		}).
		SetResult(&WxSessionRes{}).
		Get("https://api.weixin.qq.com/sns/jscode2session")
	lib.AssertErr(err)
	log.Debug().Msg(resp.String())
	err = json.Unmarshal(resp.Body(), res)
	lib.AssertErr(err)
	return res
}

// login result
type LoginRes struct {
	Name          string `json:"name"`
	StudentNumber string `json:"student_number"`
}

// 登陆 需要微信授权
func (c *UserController) PostSession() {
	// 防止重复登陆
	lib.Assert(c.Session.Get("session_key") == nil, "already_login", 401)
	// 获取请求中的code
	body := LoginReq{}
	lib.Assert(c.Ctx.ReadJSON(&body) == nil, "invalid_params", 400)
	wxRes := wxAuth(body.Code)
	log.Debug().Msg(fmt.Sprintf("code in request : %v, wxRes: %v", body.Code, wxRes))
	lib.Assert(wxRes.ErrCode == 0, wxRes.ErrMsg, 400)
	lib.Assert(c.Server.HasRegistered(wxRes.OpenId), "unregister_user", 401)
	// 维护自定义登陆状态，维护登陆状态
	c.Session.Set(IdKey, wxRes.OpenId)
	c.Session.Set("session_key", wxRes.SessionKey) // 用于构建后续的特殊请求（可能会过期）
	c.Session.Set(IdTimeKey, time.Now().Unix())
	// 构建返回信息
	userDoc := c.Server.FindUserByOpenID(wxRes.OpenId)
	lib.Assert(userDoc != nil, "unknown error")
	c.JSON(200, LoginRes{userDoc.Name, userDoc.StudentNumber})
}

// 退出登陆
func (c *UserController) DelSession() {
	lib.Assert(c.Session.Get("session_key") != nil, "not_login", 401)
	c.Session.Destroy()
	c.JSON(200)
}

type RegisterReq struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	StudentNum string `json:"student_number"`
}

// 已经授权的用户进行注册
func (c *UserController) Post() {
	body := RegisterReq{}
	lib.Assert(c.Ctx.ReadJSON(&body) == nil, "invalid_params", 400)
	// 检查用户是否注册
	wxRes := wxAuth(body.Code)
	log.Debug().Msg(fmt.Sprintf("body : %v, wxRes : %v ", body, wxRes))
	lib.Assert(wxRes.ErrCode == 0, wxRes.ErrMsg, 400)
	if OFFLINE_DEBUG {
		wxRes.OpenId = body.Code
	}
	// 防止重复注册
	lib.Assert(!c.Server.HasRegistered(wxRes.OpenId), "already_register", 400)
	c.Server.Register(body.Name, body.StudentNum, wxRes.OpenId)
	//lib.JSON(c.Ctx, 200)
	c.JSON(200)
}
