package lib

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/rs/zerolog/log"
)

type BaseRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type DataRes struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type DataListRes struct {
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Pagination Page        `json:"pagination"`
	Data       interface{} `json:"data"`
}

type Page struct {
	Page  int
	Limit int
	Total int
}

func JSON(ctx context.Context, arr ...interface{}) {
	statusCode := 200
	var page Page
	var body interface{}
	body = nil
	for _, v := range arr {
		switch v.(type) {
		case int:
			statusCode = v.(int)
		case Page:
			page = v.(Page)
		default:
			body = v
		}
	}
	var b []byte
	var err error
	if body == nil {
		b, err = jsoniter.Marshal(BaseRes{Code: statusCode, Msg: "ok"})
	} else if page.Page == 0 {
		b, err = jsoniter.Marshal(DataRes{Code: statusCode, Msg: "ok", Data: body})
	} else {
		b, err = jsoniter.Marshal(DataListRes{Code: statusCode, Msg: "ok", Data: body, Pagination: page})
	}
	Assert(err == nil, "unknown_error", iris.StatusInternalServerError)
	ctx.StatusCode(statusCode)
	if statusCode != iris.StatusNoContent {
		ctx.ContentType("application/json")
		_, err = ctx.Write(b)
		log.Debug().Msg(string(b))
		Assert(err == nil, "unknown_error", iris.StatusInternalServerError)

	}
}
