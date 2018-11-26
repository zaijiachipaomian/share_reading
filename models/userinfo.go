package models

import (
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"time"
)

// 用户的基本信息
// ID
// Phone      手机号码
// PassWord   密码
// Name       昵称
// Signature  用户的签名
// LogonTime 注册的时间
// Freeze 账户是否被系统冻结
// preMd5 密码是否经过md5
type UserInfo struct {
	ID        int64     `json:"id" orm:"pk;column(id)"`
	Phone     string    `json:"phone"`
	PassWord  string    `json:"pass_word"`
	Name      string    `json:"name"`
	Signature string    `json:"signature"`
	LogonTime time.Time `json:"logon_time"`
	Freeze    bool      `json:"-"`
	Reward    int       `json:"reward"`
	preMd5    bool
}


// 实现接口 encoding.BinaryMarshaler
func (s *UserInfo) MarshalBinary() (data []byte, err error) {

	return json.Marshal(s)
}


// 插入书籍
func (s * UserInfo) Insert() (  id int64,err error ){
	return orm.NewOrm().Insert(s )
}