package v1

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
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
	Abort400        = "400"
	Abort403        = "403"
	Abort404        = "404"
	Abort500        = "500"
	ContentType     = "Content-Type"
	ApplicationJson = "application/json"
	MinCost         = 0  // 最低花费
	MaxCost         = 20 // 最高的花费
	MinReward       = 1  // 最低的阅读奖励
	MaxReward       = 12 // 最高的阅读奖励
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
	if !strings.HasPrefix(contentType, ApplicationJson) {
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
	// 修复错误 info.ID = 返回值
	info.ID, err = info.Insert()
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
	//
	ok, sub := utils.ValidJWT(this.Ctx)
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

// ------------------------------------------------------------
// 用户发布新的书籍
type UserPublishBookController struct {
	Base
}

func (this *UserPublishBookController) Prepare() {

}

func (this *UserPublishBookController) Post() {

	ok, sub := utils.ValidJWT(this.Ctx)

	// 用户的 token 已经过期了
	if !ok {

		responseJSON(&this.Controller, models.ResponseMessage{
			Detail: "登录过期,请重新登录",
			Code:   401,
		})
		this.StopRun()
		return

	}

	var info models.UserInfo
	var err error
	err = ValidUserInfo(sub, &info)
	// 从缓存中获取用户的的消息失败
	if err != nil {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "登录信息过期,请重新登录", Code: 401})
		this.StopRun()
		return
	}
	var publishBook models.PublishBook
	// todo 获取用户的提交的基本数据, Cost  ,Reward ,
	if err := this.ParseForm(&publishBook); err != nil {
		info_(this.Ctx.Request.RemoteAddr, "  parseForm ", err)
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "参数错误", Code: 400})
		this.StopRun()
		return
	}
	fmt.Printf("publish = %+v \n", publishBook)
	// 用户提交的书籍 消耗的阅读币和 奖励币有误
	if publishBook.Cost < MinCost || publishBook.Cost > MaxCost ||
		publishBook.Reward < MinReward || publishBook.Reward > MaxReward {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "forbidden ", Code: 403})
		this.StopRun()
		return

	}

	f, h, err := this.GetFile("filename")
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

	publishBook.UserInfo = &info
	publishBook.SaveName = fmt.Sprintf("%s%d.pdf", strconv.FormatInt(time.Now().Unix(), 10), info.ID)
	publishBook.PublishTime = time.Now()
	// 保存发布书籍的信息
	_, err = publishBook.Insert()
	if err != nil {
		this.Abort(Abort500)
		return
	}

	err = this.SaveToFile("filename", "static/publish/"+publishBook.SaveName) // 保存位置在 static/upload, 没有文件夹要先创建
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
	return

}

// ---------------------------------------------------------------------
// 用户发布书单
type UserReadingListController struct {
	Base
}

func (this *UserReadingListController) Prepare() {

}

// 用作于用户创建书单
func (this *UserReadingListController) Post() {

	// 验证 jwt 是否是有效的
	ok, sub := utils.ValidJWT(this.Ctx)
	// 验证 token 失败
	if !ok {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "登录过期,请重新登录", Code: 422})
		this.StopRun()
		return
	}

	var info models.UserInfo
	var err error
	err = ValidUserInfo(sub, &info)
	// 从缓存中获取用户的的消息失败
	if err != nil {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "登录信息过期,请重新登录", Code: 401})
		this.StopRun()
		return
	}

	var readingList models.ReadingList
	// 发序列书单的基本消息
	err = deserializeJSON2Obj(&this.Controller, &readingList)
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "反序列化书单失败 ", string(this.Ctx.Input.RequestBody), err)
		this.Abort(Abort400)
		return
	}

	// 如果书单的名称为空
	if len(readingList.Name) == 0 || len(readingList.Name) >= 50 ||
		len(readingList.Types) == 0 || len(readingList.Types) >= 50 ||
		len(readingList.Instruction) == 0 || len(readingList.Instruction) >= 250 {
		info_(this.Ctx.Request.RemoteAddr, "书单的信息不正确  ", readingList)
		this.Abort(Abort403)
		return
	}

	// 设置书单的所属用户
	readingList.UserInfo = &info
	// 重置用户的ID
	readingList.Id = 0

	id, err := readingList.Insert()

	// 持久化书单失败
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "持久化书单失败 ", err, info)
		this.Abort(Abort500)
		return
	}

	readingList.Id = id
	// 返回持久化成功的书单的id
	// 客户端需要拿到这个返回的id
	responseJSON(&this.Controller, models.ResponseMessage{Detail: readingList, Code: 200})
	return

}

// ----------------------------------------------
// 用户增加书籍到书单
type UserATBook2ReadingListController struct {
	Base
}

//
func (this *UserATBook2ReadingListController) Prepare() {

	ok, sub := utils.ValidJWT(this.Ctx)

	// 用户的 token 已经过期了
	if !ok {

		responseJSON(&this.Controller, models.ResponseMessage{
			Detail: "登录过期,请重新登录",
			Code:   401,
		})
		this.StopRun()
		return

	}

	var info models.UserInfo
	var err error
	err = ValidUserInfo(sub, &info)
	// 从缓存中获取用户的的消息失败
	if err != nil {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "登录信息过期,请重新登录", Code: 401})
		this.StopRun()
		return
	}

	//  获取用户提交的书籍的信息
	var bookProfile models.BookProfile
	// 反序列化书籍的资料失败
	err = deserializeJSON2Obj(&this.Controller, &bookProfile)
	if err != nil {
		this.Abort(Abort400)
		this.StopRun()
		return
	}

	// 排除没有上传书单id 的数据
	if bookProfile.ReadingList == nil {
		info_(this.Ctx.Request.RemoteAddr, "用户添加书籍到书单   但是没有上传 书单的id")
		this.Abort(Abort400)
		this.StopRun()
		return
	}

	fmt.Printf(" bookProfile = %#v  bookList = %+v \n ", bookProfile, bookProfile.ReadingList)

	// 设置书单的用户信息
	bookProfile.ReadingList.UserInfo = &info
	fmt.Println(info)
	// 查找书单的信息 , 查看用户该书单是否存在
	err = orm.NewOrm().Read(bookProfile.ReadingList, "UserInfo", "Id")

	if err != nil {
		info_(this.Ctx.Request.RemoteAddr, "用户添加图书到书单,但是找不到该书单的信息", err)
		this.Abort(Abort404)
		this.StopRun()
	}

	bookProfile.Id, err = bookProfile.Insert()
	if err != nil {
		this.Abort(Abort500)
		this.StopRun()
		return
	}

	responseJSON(&this.Controller, models.ResponseMessage{Detail: bookProfile, Code: 200})

}

// 用户增加书籍到指定的书单
func (this *UserATBook2ReadingListController) Post() {

}

// ---------------------------------------------------
// 用户注销登录
type UserLogoutController struct {
	Base
}

// 用户注销登录, 删除token .uuids 做无效处理
func (this *UserLogoutController) Get() {

	ok, sub := utils.ValidJWT(this.Ctx)

	if !ok {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "ok", Code: 200})
		this.StopRun()
		return
	}

	utils.GetClient().Del(sub).Result()

	responseJSON(&this.Controller, models.ResponseMessage{Detail: "ok", Code: 200})
	this.StopRun()
	return

}

// --------------------------------------------
// 用户评价书籍
type UserCommentBookController struct {
	Base
}

func (this *UserCommentBookController) Post() {
	pd := struct {
		BookId  int64  `json:"book_id"`
		Content string `json:"content"`
	}{}

	ok, sub := utils.ValidJWT(this.Ctx)

	if !ok {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "用户登录过期,请重新登录", Code: 200})
		this.StopRun()
		return
	}

	var info models.UserInfo
	var err error
	err = ValidUserInfo(sub, &info)
	if err != nil {
		responseJSON(&this.Controller, models.ResponseMessage{Detail: "用户登录过期,请重新登录", Code: 200})
		this.StopRun()
		return
	}

	// 反序列化用户提交的数据信息
	// 带评论书籍的id
	// 评价的内容
	err = deserializeJSON2Obj(&this.Controller, &pd)
	if err != nil  || len(pd.Content) >=250{
		info_(this.Ctx.Request.RemoteAddr,"反序列反消息失败或者评价的内容过长 ",pd, " err ", err )
		this.Abort(Abort400)
		this.StopRun()
		return
	}

	// 如果评级的内容为空
	if len(pd.Content) == 0{
		info_(this.Ctx.Request.RemoteAddr, "禁止空白的评论")
		this.Ctx.ResponseWriter.WriteHeader(403)
		this.Ctx.WriteString(`{"detail":"禁止无效的评论","code":403}`)
		this.StopRun()
		return
	}

	var bookProfile models.BookProfile
	bookProfile.Id = pd.BookId
	err = orm.NewOrm().Read(&bookProfile, "id")

	if err != nil {
		// 该书籍可能不存在
		info_(this.Ctx.Request.RemoteAddr, " 查找书籍的信息 "  , err)
		this.Abort(Abort500)
		return
	}

	var bookComment models.BookComment

	// 初始化 评论的基本内容
	bookComment.Content = pd.Content
	bookComment.UserInfo = &info
	bookComment.BookProfile  = &  bookProfile
	bookComment.CommentTime = time.Now()

	bookComment.Id ,err = bookComment.Insert()
	if err != nil {
		info_(this.Ctx.Request.RemoteAddr,"持久化书籍评论失败 " ,err , info )
		this.Abort(Abort500)
		return
	}

	// 评论成功, 返回
	responseJSON(&this.Controller, models.ResponseMessage{Detail:"succeed",Code:200})
	return


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

// 给客户端返回 JSON格式的字符串
func responseJSON(ctr *beego.Controller, v interface{}) {
	ctr.Data[controllers.DataJson] = v
	ctr.ServeJSON(true)
}

// 从缓存中获取用户的数据
func ValidUserInfo(uuids string, info *models.UserInfo) (err error) {
	bytes, err := utils.GetClient().Get(uuids).Bytes()

	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, info)
	if err != nil {
		return
	}
	return
}

// info 打印日志消息
func info_(f interface{}, v ... interface{}) {
	logs.Info(f, v...)
}
