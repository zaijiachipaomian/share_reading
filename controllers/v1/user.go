package v1

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"math/rand"
	"neecola.com/eula/controllers"
	"reading/models"
	"reading/utils"
	"strconv"
	"strings"
	"time"
)

// ------------------------------------------------
// 用户登录的控制器
type UserLoginController struct {
	Base
}

const (
	Abort400       = "400"
	Abort403       = "403"
	Abort404       = "404"
	Abort500       = "500"
	ContentType    = "Content-Type"
	ApplicatonJson = "application/json"
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
		info_(this.Ctx.Request.RemoteAddr, " 用户登录 ", pd.Phone, " 用户密码 ", pd.PassWord)
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

	//设置过期时间
	//使用 指针的形式
	_, err = utils.GetClient().Set(uuids.String(), &info, (time.Hour)*24*30*12).Result()

	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "set user info redis err", err)
		this.Abort(Abort500)
		return
	}
	var infos models.UserInfo

	da, err := utils.GetClient().Get(uuids.String()).Bytes()
	if err != nil {
		fmt.Println("get da err ", err)
	}
	err = json.Unmarshal(da, &infos)
	fmt.Println("err un err ", err, " da = ", string(da))

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

type UserRegisterController struct {
	Base
}

// 预处理
// 过滤不是 不是JSON 方式提交的数据

func (this *UserRegisterController) Prepare() {
	contentType := this.Ctx.Request.Header.Get(ContentType)
	//过滤Content-Type 不是application/json方式的请求
	if !strings.HasPrefix(contentType, ApplicatonJson) {
		//
		info_(this.Ctx.Request.RemoteAddr, " content -type 不是符合的方式 ", contentType)
		this.Abort(Abort400)
		return
	}

}

// 用户获取注册验证
func (this *UserRegisterController) PullValidCode() {
	phone := struct {
		Phone string `json:"phone"`
	}{}

	err := deserializeJSON2Obj(&this.Controller, &phone)

	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, " 反序列化获取手机号码失败 ", err)
		this.Abort(Abort400)
		return
	}

	ok, err := utils.RegexpValidPhone(phone.Phone, utils.PhonePattern)

	if !ok {
		info_(this.Ctx.Request.RemoteAddr, " 提交手机号码有误 ", err)
		this.Abort(Abort403)
		return
	}

	data, err := utils.GetClient().Get(phone.Phone).Bytes()

	// 该用户的手机号码已经被使用
	if len(data) != 0 {
		this.Data[controllers.DataJson] = models.ResponseMessage{
			Detail: "对不起, 该手机号码无法使用",

			Code: 422,
		}
		// this.Ctx.ResponseWriter.WriteHeader(422)

		this.ServeJSON(true)
		this.StopRun()
	}

	fmt.Println("s = ", string(data), " err ", err)

	// 生成验证码
	validCode := fmt.Sprintf("%6d", rand.Int31n(999999))

	fmt.Println("验证码 ", validCode)

	// 修正保存的是手机号码
	_, err = utils.GetClient().Set(fmt.Sprintf("%s:%s", phone.Phone, validCode), validCode, time.Minute*5).Result()

	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "  保存手机的验证码失败 ", err)
		this.Abort(Abort500)
		return
	}
	// todo
	// 发送验证码

	this.Data[controllers.DataJson] = models.ResponseMessage{
		Detail: "验证码已经发送,请注意查看短信,",
		Code:   200,
	}
	this.ServeJSON(true)
}

// 用户注册
// 验证用户提交的验证码
// 验证 手机号码
func (this *UserRegisterController) Register() {
	dp := struct {
		Phone     string `json:"phone"`
		PassWord  string `json:"pass_word"`
		ValidCode string `json:"valid_code"`
	}{}

	err := deserializeJSON2Obj(&this.Controller, &dp)

	// 提交数据的格式有错
	// JSON 格式发序列化失败
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, " deserializeJson2Obj err ", err)
		this.Abort(Abort400)
		return
	}

	// 正则表达式验证手机号码
	ok, err := utils.RegexpValidPhone(dp.Phone, utils.PhonePattern)
	// 手机号码不正确
	if !ok {
		info_(this.Ctx.Request.RemoteAddr, "手机号码不符合 ", err)
		this.Abort(Abort403)
		return
	}
	// 密码 或者 验证码不正确
	if len(dp.PassWord) < 8 || len(dp.ValidCode) != 6 {
		info_(this.Ctx.Request.RemoteAddr, " 密码长度获取 验证码长度不正确", dp.PassWord, "   ", dp.ValidCode)
		this.Abort(Abort400)
		return
	}

	val := utils.GetClient().Get(fmt.Sprintf("%s:%s", dp.Phone, dp.ValidCode)).Val()

	if val != dp.ValidCode {
		info_(this.Ctx.Request.RemoteAddr, " 验证码不正确或者已经过期", val, "  ", dp.ValidCode)
		this.Abort(Abort403)
		return
	}

	var info models.UserInfo

	info.Phone = dp.Phone
	info.PassWord = dp.PassWord
	info.LogonTime = time.Now()

	// 持久化到数据库
	_, err = info.Insert()
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, " 持久化数据库失败 ", err)
		this.Abort(Abort500)
		return

	}

	// 生成uuid
	uuids, err := uuid.NewV4()
	// 生成 uuid失败
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, " uuid.NewV4 ", " err ", err)
		this.Abort(Abort500)
		return
	}

	jswt := generateJWT(uuids)
	// 设置 auth  响应头消息
	this.Ctx.ResponseWriter.Header().Set("Authorization", jswt)

	_, err = utils.GetClient().Set(uuids.String(), &info, (time.Hour)*24*30*12).Result()
	utils.GetClient().Set(info.Phone, &info, (time.Hour)*24*30*12).Result()

	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "set user info redis err", err)
		this.Abort(Abort500)
		return
	}

	// 查找相关的数据
	// todo
	this.Data["json"] = models.ResponseMessage{
		Detail: "pass",
		Code:   200,
	}
	this.ServeJSON(true)

	fmt.Println(" val  ", val)

}

//------------------------------------ 用户上传书籍
type UserUploadController struct {
	Base
}

func (this *UserUploadController) Prepare() {

}

func (this *UserUploadController) Post() {
	ok, sub := validJWT(&this.Controller)
	if !ok || sub == "" {
		this.Data[controllers.DataJson] = models.ResponseMessage{
			Detail: "登录过期,请重新登录",
			Code:   403,
		}
		this.ServeJSON(true)
		this.StopRun()
	}
	var info models.UserInfo
	data, err := utils.GetClient().Get(sub).Bytes()

	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "  upload file ", err)
		this.Abort(Abort403)
		return
	}
	err = json.Unmarshal(data, &info)

	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "  upload file ", err)
		this.Abort(Abort403)
		return
		this.Abort(Abort403)

	}
	f, h, err := this.GetFile("uploadname")
	if err != nil {
		info_("getfile err ", err)
		this.Abort(Abort403)
		return
	}
	defer f.Close()

	if !strings.HasSuffix(h.Filename, "pdf") {
		info_(this.Ctx.Request.RemoteAddr, "不允许上传非 pdf 文件 ", )
		this.Data[controllers.DataJson] = models.ResponseMessage{
			Detail: "不允许上传非pdf文件",
			Code:   422,
		}
		this.ServeJSON(true)
		return
	}

	book := models.UploadBook{}
	book.UserInfo = &info
	book.BookName = h.Filename
	book.UploadTime = time.Now()
	book.SaveName = fmt.Sprintf("%s%d.pdf", strconv.FormatInt(time.Now().Unix(), 10), info.ID)
	_, err = book.Insert()
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, " 保存信息失败", err)
		this.Abort(Abort500)
		return
	}
	err = this.SaveToFile("uploadname", "static/upload/"+book.SaveName) // 保存位置在 static/upload, 没有文件夹要先创建
	// 返回处理
	if err != nil {
		this.Data[controllers.DataJson] = models.ResponseMessage{
			Detail: "上传错误",
			Code:   422,
		}
		this.ServeJSON(true)
		return
	}

	this.Data[controllers.DataJson] = models.ResponseMessage{
		Detail: "上传书籍成功",
		Code:   200,
	}
	this.ServeJSON(true)

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

func validJWT(ctr *beego.Controller) (ok bool, sub string) {

	authorization := ctr.Ctx.Request.Header.Get("Authorization")

	if authorization == "" {
		return ok, ""
	}

	t, err := jwt.Parse(authorization, func(*jwt.Token) (interface{}, error) {
		return jwtSigningKey, nil
	})

	if err != nil {
		return ok, ""
	}

	iss, ok := t.Claims.(jwt.MapClaims)
	if ok {
		fmt.Printf("s = %+v \n", iss["sub"])
		sub = iss["sub"].(string)
	} else {
		fmt.Printf("error t.cliams = %#v \n", t.Claims)
	}

	ok = t.Valid
	return ok, sub
}
