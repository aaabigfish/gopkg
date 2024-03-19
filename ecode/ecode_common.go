package ecode

// All common ecode [1000001-1000999]
var (
	EcodeOk = New(200) // 成功

	InvalidParam = New(1000001) // 参数错误
	NotLogin     = New(1000002) // 没有登录
	SignCheckErr = New(1000003) // 检查签名错误
	NotFound     = New(1000004) // 没有找到
	Forbidden    = New(1000005) // 没有权限
)
