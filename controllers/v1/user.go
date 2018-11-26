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
	Abort403 = "403"
	Abort404 = "404"
	Abort500 = "500"
)

var (
	jwtSigningKey = []byte("bla bla bla")
)

func init() {
	var i models.UserInfo
	i.Phone = "13123456789"
	i.PassWord = "12345678"
	utils.GetClient().Set(i.Phone, &i, -1).Result()
}

// 用户登录
func (this *UserLoginController) Post() {
	// 用来装 上传的手机号码和密码
	pd := struct {
		Phone    string `json:"phone"`
		PassWord string `json:"pass_word"`
	}{}
	//var info models.UserInfo
	// 反序列化提交的数据
	err := deserializeJSON2Obj(&this.Controller, &pd)
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, err)
		this.Abort(Abort400)
	}

	// 正则表达式验证手机号码
	// @ pattern 正则表达式的类型
	// @ phone   待验证的字符串
	ok, err := utils.RegexpValidPhone(pd.Phone, utils.PhonePattern)
	// 如果手机号码不正确, 或者手机号码的长度小于8
	// 返回提交的数据不正确
	if !ok || len(pd.PassWord) < 8 {
		info_(this.Ctx.Request.RemoteAddr, err, )
		this.Abort(Abort400)
	}

	// 匹配用户名和密码
	val := utils.GetClient().Get(pd.Phone).Val()

	// val == "" 表示用户获取的缓存中没有这个数据
	// 表示用户没有注册这个手机号码没有被注册
	if val == "" {
		this.Abort(Abort404)
		return
	}
	// 反序列化 用户的数据
	var info models.UserInfo
	err = json.Unmarshal([]byte(val), &info)

	// 反序列化数据失败
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, err)
		this.Abort(Abort500)
		return
	}

	fmt.Printf("info = %+v \n", info)
	// 用户的密码不正确
	if pd.PassWord != info.PassWord {
		this.Abort(Abort404)
		return
	}

	// 该用户已经被冻结了
	if info.Freeze == true {
		this.Abort(Abort403)
		return
	}

	fmt.Println("info ", info)
	fmt.Println("val = " + val)
	// 确定用户是否已经被冻结

	// 查找用户是否已经登录

	// 生成 uuid
	// 生成uuid
	uuids, err := uuid.NewV4()
	// 生成 uuid失败
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, " uuid.NewV4 ", " err ", err)
		this.Abort(Abort500)
		return
	}

	// 设置过期时间
	// 使用 指针的形式
	//_, err = utils.GetClient().Set(uuids.String(), &info, (time.Hour)*24*30*12).Result()
	//
	//if err != nil {
	//	info_(this.Ctx.Request.RemoteAddr, "set user info redis err", err)
	//	this.Abort(Abort500)
	//	return
	//}
	var infos models.UserInfo

	da, err  := utils.GetClient().Get(uuids.String()).Bytes()
	if err != nil {
		fmt.Println("get da err ",err )
	}
	err  = json.Unmarshal(da, &infos)
	fmt.Println("err un err ",err , " da = ", string(da))

	fmt.Printf("infos = %+v\n", infos)

	// 生成 jwt
	jswt := generateJWT(uuids)
	// 设置 auth  响应头消息
	this.Ctx.ResponseWriter.Header().Set("Authorization", jswt)

	// 查找相关的数据
	// todo
	this.Data["json"] = models.ResponseMessage{
		Detail: "pass",
		Code:   200,
	}
	this.ServeJSON(true)

	//t, err := jwt.Parse(jswt, func(*jwt.Token) (interface{}, error) {
	//	return jwtSigningKey, nil
	//})
	//
	//if err != nil {
	//	fmt.Printf("jwt.Parse error %+v \n", err)
	//	this.Abort(Abort500)
	//	return
	//}
	//
	//iss, ok := t.Claims.(jwt.MapClaims)
	//if ok {
	//	fmt.Printf("s = %+v \n", iss["sub"])
	//} else {
	//	fmt.Printf("error t.cliams = %#v \n", t.Claims)
	//}

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
		ExpiresAt: int64(time.Now().Unix() + int64(time.Hour)*24*30*12),
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
func info_(f interface{}, v ... interface{}) {
	logs.Info(f, v...)
}
