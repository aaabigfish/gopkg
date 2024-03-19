package ginx

import "fmt"

const (
	ResultFail = 0
	ResultOk   = 200
)

const (
	ErrorCodeSuccess         = "200"
	ErrorCodeFail            = "0"
	ErrorCodeInvalidArgument = "400"
	ErrorCodeUnauthenticated = "401"
)

const (
	ErrorMsgSuccess         = "OK"
	ErrorMsgFail            = "ERROR"
	ErrorMsgInvalidArgument = "InvalidArgument"
	ErrorMsgNotLogin        = "Unauthenticated"
)

type M map[string]interface{}

type Result struct {
	ResultCode int         `json:"code"`
	ErrorCode  string      `json:"error_code,omitempty"`
	ErrorMsg   string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

type PageInfo struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	List  interface{} `json:"list"`
}

type Option func(*Result)

func ResultCode(resultCode int) Option {
	return func(ret *Result) {
		ret.ResultCode = resultCode
	}
}

func ErrorCode(errorCode string) Option {
	return func(ret *Result) {
		ret.ErrorCode = errorCode
	}
}

func ErrorMsg(errorMsg string) Option {
	return func(ret *Result) {
		ret.ErrorMsg = errorMsg
	}
}

func Data(data interface{}) Option {
	return func(ret *Result) {
		ret.Data = data
	}
}

func New(options ...Option) *Result {
	ret := &Result{}

	for _, opt := range options {
		opt(ret)
	}

	return ret
}

// 1:error_msg 2:error_code
func NewFailData(data interface{}, errs ...string) *Result {
	ret := &Result{
		ResultCode: ResultFail,
		ErrorCode:  ErrorCodeFail,
		ErrorMsg:   ErrorMsgFail,
	}

	ret.Data = data

	for i, err := range errs {
		if i == 0 { // error_msg
			ret.ErrorMsg = err
		} else if i == 1 { // error_code
			ret.ErrorCode = err
		} else {
			break
		}
	}

	return ret
}

// errs 1:error_msg 2:error_code
func NewFail(code int, data interface{}, errs ...string) *Result {
	ret := &Result{
		ResultCode: code,
		ErrorCode:  ErrorCodeFail,
		ErrorMsg:   ErrorMsgFail,
	}

	ret.Data = data

	for i, err := range errs {
		if i == 0 { // error_msg
			ret.ErrorMsg = err
		} else if i == 1 { // error_code
			ret.ErrorCode = err
		} else {
			break
		}
	}

	return ret
}

// errs 1:error_msg 2:error_code
func NewFailMsg(errs ...string) *Result {
	ret := &Result{
		ResultCode: ResultFail,
		ErrorCode:  ErrorCodeFail,
		ErrorMsg:   ErrorMsgFail,
	}

	for i, err := range errs {
		if i == 0 { // error_msg
			ret.ErrorMsg = err
		} else if i == 1 { // error_code
			ret.ErrorCode = err
		} else {
			break
		}
	}

	return ret
}

// 1:data 2:error_msg 3:error_code
func NewOk(data interface{}, errs ...string) *Result {
	ret := &Result{
		ResultCode: ResultOk,
		ErrorCode:  ErrorCodeSuccess,
		ErrorMsg:   ErrorMsgSuccess,
	}

	ret.Data = data

	for i, err := range errs {
		if i == 0 { // error_msg
			ret.ErrorMsg = err
		} else if i == 1 { // error_code
			ret.ErrorCode = err
		} else {
			break
		}
	}

	return ret
}

// 1:PageInfo 2:error_msg 3:error_code
func Page(page *PageInfo, errs ...string) *Result {
	ret := &Result{
		ResultCode: ResultOk,
		ErrorCode:  ErrorCodeSuccess,
		ErrorMsg:   ErrorMsgSuccess,
	}

	ret.Data = page

	for i, err := range errs {
		if i == 0 { // error_msg
			ret.ErrorMsg = err
		} else if i == 1 { // error_code
			ret.ErrorCode = err
		} else {
			break
		}
	}

	return ret
}

func InvalidArgumentError() *Result {
	ret := &Result{
		ResultCode: ResultFail,
		ErrorCode:  ErrorCodeInvalidArgument,
		ErrorMsg:   ErrorMsgInvalidArgument,
	}

	return ret
}

func NotLoginError() *Result {
	ret := &Result{
		ResultCode: ResultFail,
		ErrorCode:  ErrorCodeUnauthenticated,
		ErrorMsg:   ErrorMsgNotLogin,
	}

	return ret
}

func (s *Result) String() string {
	return fmt.Sprintf("ResultCode:[%d] ErrorCode:[%s]%s", s.ResultCode, s.ErrorCode, s.ErrorMsg)
}
