package libx

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Uid(c *gin.Context) uint {
	uid := c.MustGet("uid").(uint)
	uidInt := uid
	return uidInt
}

func GetUsername(c *gin.Context) string {
	username := c.MustGet("username").(string)
	return username
}

// HTTPStatusKey 与 ResponseMiddleware 约定：只写入 context，不要调用 c.Status。
// 若此处调用 c.Status，会触发 WriteHeader，导致 Gin 认为响应已写出，统一 JSON 中间件会跳过，客户端收到空 body。
const HTTPStatusKey = "http_status"

func Code(c *gin.Context, code int) {
	c.Set(HTTPStatusKey, code)
}

func Msg(c *gin.Context, msg string) {
	c.Set("message", msg)
}

func Data(c *gin.Context, data interface{}) {
	c.Set("data", data)
}

func Err(c *gin.Context, code any, msg string, err error) {
	codeInt, ok := code.(int)
	if ok {
		Code(c, codeInt)
	} else {
		Code(c, 500)
	}

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}
	c.Set("code", code)
	c.Set("message", msg+" "+errorMsg)
	// 打印错误信息
	log.Println(msg + " " + errorMsg)
}

// Ok 一个参数省略msg
func Ok(c *gin.Context, input ...interface{}) {
	if len(input) >= 3 {
		log.Println("too many parameters")
		Err(c, 500, "参数过多，请后端开发人员排查", nil)
	}
	Code(c, 200)
	if len(input) == 2 {
		Msg(c, input[0].(string))
		Data(c, input[1])
	} else {
		Msg(c, input[0].(string))
		Data(c, nil)
	}
}
