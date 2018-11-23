package v1

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"reading/models"
	"reading/utils"
	"time"
)

// ------------------------------------------------
// 用户登录的控制器
type UserLoginController struct {
	Base
}

const (
	Abort400 = "400"
	Abort500 = "500"

)

var(
	jwtSigningKey = []byte("bla bla bla")
)


// 用户登录
func (this *UserLoginController) Post() {
	var info models.UserInfo
	// 反序列化提交的数据
	err := deserializeJSON2Obj(&this.Controller, &info)
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr,err )
		this.Abort(Abort400)
	}

	// 正则表达式验证手机号码
	// @ pattern 正则表达式的类型
	// @ phone   待验证的字符串
	ok, err := utils.RegexpValidPhone(info.Phone, utils.PhonePattern)
	// 如果手机号码不正确, 或者手机号码的长度小于8
	// 返回提交的数据不正确
	if !ok || len(info.PassWord) < 8 {
		info_(this.Ctx.Request.RemoteAddr,err,)
		this.Abort(Abort400)
	}

	// 匹配用户名和密码

	// 确定用户是否已经被冻结

	// 查找用户是否已经登录

	// 生成 uuid
	// 生成uuid
	uuids, err := uuid.NewV4()
	if err != nil {
		this.Abort(Abort500)
	}

	jswt := generateJWT(uuids)


	t, err := jwt.Parse(jswt, func(*jwt.Token) (interface{}, error) {
		return jwtSigningKey, nil
	})


	if err != nil {
		fmt.Printf("jwt.Parse error %+v \n", err )
		this.Abort(Abort500)
	}

	iss,ok  := t.Claims.(jwt.MapClaims)
	if ok {
		fmt.Printf("s = %+v \n",iss["sub"] )
	} else {
		fmt.Printf("error t.cliams = %#v \n", t.Claims)
	}


	// 生成 jwt
	// 返回 Auth
	// 返回用户的相关数据

}

// 从提交的数据中使用json反序列化到v
func deserializeJSON2Obj(ctr *beego.Controller, v interface{}) (err error) {
	err = json.Unmarshal(ctr.Ctx.Input.RequestBody, v)
	return err
}

// 生成jwt
func generateJWT(uuids uuid.UUID) (jswt string) {

	var err error
	claims := jwt.StandardClaims{
		NotBefore: int64(time.Now().Unix() - 1000),
		// 过期时间设置为一年
		ExpiresAt: int64(time.Now().Unix() +int64( time.Hour) * 24 * 30 * 12 ),
		Issuer:    "reading",
		Subject:   uuids.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jswt, err = token.SignedString(jwtSigningKey)
	if err != nil {
		fmt.Println("err > ", err.Error())
		return jswt
	}

	return jswt
}
// info 打印日志消息
func info_(f interface{} , v ... interface{}){
	logs.Info(f, v...)
}