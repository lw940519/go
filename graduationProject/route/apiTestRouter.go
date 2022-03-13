package route

import (
	"github.com/fasthttp/router"
)

// APITestRouter 路由
func APITestRouter() *router.Router {
	r := router.New()
	r.GET("/run", test.Run) // 健康检查

	return r
}
