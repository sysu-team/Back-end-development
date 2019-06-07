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
	b.Handle("GET", "/me", "GetMe", withLogin)

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
			"appid":      WxAppid,
			"secret":     WxSecret,
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

// 登陆 需要微信授权
func (c *UserController) PostSession() {
	// 防止重复登陆
	lib.Assert(c.Session.Get(WxSessionKey) == nil, "already_login", 401)
	// 获取请求中的code
	body := LoginReq{}
	lib.Assert(c.Ctx.ReadJSON(&body) == nil, "invalid_params", 400)
	wxRes := wxAuth(body.Code)
	log.Debug().Msg(fmt.Sprintf("code in request : %v, wxRes: %v", body.Code, wxRes))
	lib.Assert(wxRes.ErrCode == 0, wxRes.ErrMsg, 400)
	lib.Assert(c.Server.HasRegistered(wxRes.OpenId), "unregister_user", 401)
	// 维护自定义登陆状态，维护登陆状态
	log.Debug().Msg("session id : " + c.Session.ID())
	c.Session.Set(IdKey, wxRes.OpenId)
	c.Session.Set(WxSessionKey, wxRes.SessionKey) // 用于构建后续的特殊请求（可能会过期）
	c.Session.Set(IdTimeKey, time.Now().Unix())
	// 构建返回信息
	c.JSON(200, c.Server.GetUserInfo(wxRes.OpenId))
}

// 退出登陆
func (c *UserController) DelSession() {
	lib.Assert(c.Session.Get(IdKey) != nil, "not_login", 401)
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
	lib.Assert(!c.Server.HasRegistered(wxRes.OpenId), "duplicated_username", 401)
	lib.Assert(!c.Server.HasRegistered(body.StudentNum), "duplicated_student_num", 402)
	c.Server.Register(body.Name, body.StudentNum, wxRes.OpenId)
	//lib.JSON(c.Ctx, 200)
	c.JSON(200)
}

//  已经登陆的用户获取用户信息
func (c *UserController) GetMe() {
	c.JSON(200, c.Server.GetUserInfo(c.Session.GetString(IdKey)))
}
