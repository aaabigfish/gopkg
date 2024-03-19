package ecode

// All notify ecode [2001001-2001999]

var (
	NotifySubmitFail   = New(2001001) //提交失败
	NotifyMegTooLong   = New(2001002) //通知消息长度超过限制
	NotifyTargetUrlErr = New(2001003) //通知消息url错误
	NotifyMethodErr    = New(2001004) //通知消息method错误
	NotifyTitleErr     = New(2001005) //通知消息title错误
)
