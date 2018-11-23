// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"github.com/astaxie/beego"
	"reading/controllers"
	"reading/controllers/v1"
)

func init() {
	ver1 := beego.NewNamespace("/v1",
		beego.NSNamespace("/user",
			// 用户登录的路径
			beego.NSRouter("/login", &v1.UserLoginController{},"post:Post"),
		),
	)

	// 添加路由
	beego.AddNamespace(ver1)

	// 处理出错的路由
	beego.ErrorController(&controllers.ErrorController{})
}
