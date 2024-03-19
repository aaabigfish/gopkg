package ginx

import "github.com/gin-gonic/gin"

func Ok(c *gin.Context, data interface{}, errs ...string) {
	c.JSON(200, NewOk(data, errs...))
}

func Fail(c *gin.Context, code int, data interface{}, errs ...string) {
	c.JSON(200, NewFail(code, data, errs...))
}

func FailMsg(c *gin.Context, errs ...string) {
	c.JSON(200, NewFailMsg(errs...))
}

func FailData(c *gin.Context, data interface{}, errs ...string) {
	c.JSON(200, NewFailData(data, errs...))
}
