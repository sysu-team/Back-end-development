package lib

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris/context"
	"github.com/rs/zerolog/log"
)

type ErrorRes struct {
	Code int
	Msg  string
}

// Assert Web断言，产生的 panic 经由调用 chain 传播到 注册得中间件中得 error handler 中 recover 中，进行统一处理
func Assert(condition bool, msg string, code ...int) {
	if !condition {
		statusCode := 400
		if len(code) > 0 {
			statusCode = code[0]
		}
		panic(ErrorRes{Code: statusCode, Msg: msg})
	}
}

// AssertErr error断言
func AssertErr(err error, code ...int) {
	if err != nil {
		log.Error().Msg(err.Error())
		statusCode := 400
		if len(code) > 0 {
			statusCode = code[0]
		}
		panic(ErrorRes{Code: statusCode, Msg: err.Error()})
	}
}

// NewErrorHandler 构造错误处理Handler
func NewErrorHandler() context.Handler {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}

				switch err.(type) {
				case ErrorRes:
					res := err.(ErrorRes)
					ctx.StatusCode(res.Code)
					ctx.ContentType("application/json")
					b, e := jsoniter.Marshal(res)
					if e != nil {
						break
					}
					_, err = ctx.Write(b)
					if err == nil && res.Code < 500 {
						return
					}
				}
				panic(err)
			}
		}()

		ctx.Next()
	}
}
