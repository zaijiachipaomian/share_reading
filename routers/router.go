// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"reading/controllers"
	"reading/controllers/v1"
	"reading/models"
	"reading/utils"
)

func init() {
	ver1 := beego.NewNamespace("/v1",
		beego.NSNamespace("/user",
			// 用户登录的路径
			beego.NSRouter("/login", &v1.UserLoginController{}, "post:Post"),

			// 请求注册验证码
			beego.NSRouter("/register/pull_valid", &v1.UserRegisterController{}, "post:PullValidCode"),

			// 用户注册
			beego.NSRouter("/register", &v1.UserRegisterController{}, "post:Register"),

			// 用户上传自己的文件
			beego.NSRouter("/upload/self", &v1.UserUploadController{}, ),

			// 用户发布书籍
			beego.NSRouter("/publish", &v1.UserPublishBookController{}, "post:Post"),

			// 用户发布书单
			beego.NSRouter("/p/reading_list", &v1.UserReadingListController{} , "post:Post"),

			// 用户添加书籍到书单
			// 如果书单的id at = add to
			beego.NSRouter("/at/reading_list", &v1.UserATBook2ReadingListController{}, "post:Post"),

			// 用户注销的功能
			beego.NSRouter("/logout",&v1.UserLogoutController{}),
		),

	)

	// 添加路由
	beego.AddNamespace(ver1)

	// 过滤静态数据的请求
	// 用户下载静态文件的过滤请求
	beego.InsertFilter("/static/upload/:SaveName([0-9]+).pdf", beego.BeforeStatic, func(ctx *context.Context) {
		// 获取文件名
		saveName := ctx.Input.Param(":SaveName")
		logs.Info(ctx.Request.RemoteAddr, " dl "+saveName)

		// 验证 jwt
		ok, sub := utils.ValidJWT(ctx)

		if !ok {
			// 无效 令牌, 请重新登录
			// 或者说用户没有携带令牌的参数
			ctx.ResponseWriter.WriteHeader(422)
			res := struct {
				Detail interface{} `json:"detail"`
				Code   int         `json:"code"`
			}{
				Detail: "登录过期, 请重新登录",
				Code:   422,
			}

			str, _ := utils.Marshal2JSONString(res)

			ctx.WriteString(str)
			return
		}

		// 验证用户是否上传这本书籍作为保存资料
		btd, err := utils.GetClient().Get(sub).Bytes();
		if err != nil {
			ctx.ResponseWriter.WriteHeader(500)
			ctx.WriteString("内部错误")
			return
		}
		// 如果查找不到和该和用户相关的信息
		if len(btd) == 0 {
			ctx.ResponseWriter.WriteHeader(404)
			ctx.WriteString("用户登录过期,请重新登录")
			return
		}

		var info models.UserInfo

		// 反序列 用户的数据
		err = json.Unmarshal(btd, &info)
		if err != nil {
			ctx.ResponseWriter.WriteHeader(500)
			ctx.WriteString("内部错误")
			return
		}

		book := models.UploadBook{}
		book.UserInfo = &info
		book.SaveName = saveName + ".pdf"
		// 查找用户是否上传这一本书籍
		err = orm.NewOrm().Read(&book, "UserInfo", "SaveName")
		if err != nil {
			logs.Info(ctx.Request.RemoteAddr, " 查找数据库错误 ", err)
			ctx.ResponseWriter.WriteHeader(403)
			ctx.WriteString("不允许访问")
			return
		}

	})

	// 用户下载其他用户发布的书籍
	// 过滤书籍是否是违法的书籍
	// 过滤下载书籍是否已经登录. 或者说用户的登录已经过期
	// 判断书籍的版权 是否存在问题,
	// 判断 书籍是需要否收费
	// 判断用户的奖励金是否足够,扣除下载本次书籍的操作
	beego.InsertFilter("/static/publish/:SaveName([0-9]+).pdf", beego.BeforeStatic, func(ctx *context.Context) {

		// 获取文件名
		saveName := ctx.Input.Param(":SaveName")
		logs.Info(ctx.Request.RemoteAddr, " dl "+saveName)

		// 验证 jwt
		ok, sub := utils.ValidJWT(ctx)

		if !ok {
			// 无效 令牌, 请重新登录
			// 或者说用户没有携带令牌的参数
			ctx.ResponseWriter.WriteHeader(422)
			res := struct {
				Detail interface{} `json:"detail"`
				Code   int         `json:"code"`
			}{
				Detail: "登录过期, 请重新登录",
				Code:   422,
			}

			str, _ := utils.Marshal2JSONString(res)

			ctx.WriteString(str)
			return
		}

		// 验证用户是否上传这本书籍作为保存资料
		btd, err := utils.GetClient().Get(sub).Bytes();
		if err != nil {
			ctx.ResponseWriter.WriteHeader(500)
			ctx.WriteString("内部错误")
			return
		}
		// 如果查找不到和该和用户相关的信息
		if len(btd) == 0 {
			ctx.ResponseWriter.WriteHeader(404)
			ctx.WriteString("用户登录过期,请重新登录")
			return
		}

		var info models.UserInfo

		// 反序列 用户的数据
		err = json.Unmarshal(btd, &info)
		if err != nil {
			ctx.ResponseWriter.WriteHeader(500)
			ctx.WriteString("内部错误")
			return
		}

		book := models.PublishBook{}
		book.SaveName = saveName + ".pdf"
		err = orm.NewOrm().Read(&book, "SaveName")
		if err != nil {
			logs.Info(ctx.Request.RemoteAddr, " 查找数据库错误 ", err)
			ctx.ResponseWriter.WriteHeader(403)
			ctx.WriteString("不允许访问")
			return
		}

		// 如果发布的书籍不是公开的,  内容非法, 版权 不正确 不允许用户 进行下载的操作
		if !book.Expose || !book.ContentIllegal || !book.CopyRight {
			logs.Info(ctx.Request.RemoteAddr, "下载书籍 书籍不允许下载,可能是权限不足", )
			ctx.ResponseWriter.WriteHeader(403)
			ctx.WriteString("不允许访问")
			return
		}

		//fmt.Printf("book = %+v  info = %+v \n"  , book, book.UserInfo)
		// todo
		//更新书籍库,, 跟新缓存 (更新用户的 奖励金的情况)
		info.Reward -= book.Cost

	})
	// 处理出错的路由
	beego.ErrorController(&controllers.ErrorController{})
}
