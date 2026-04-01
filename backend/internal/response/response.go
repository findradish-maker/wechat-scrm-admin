package response

import "github.com/gin-gonic/gin"

func JSON(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

func Success(c *gin.Context, data interface{}) {
	JSON(c, 0, "ok", data)
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
