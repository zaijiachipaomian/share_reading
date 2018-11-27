package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type PublishBook struct {
	Id             int64     `json:"id" orm:"pk"`             // 书籍的id
	UserInfo       *UserInfo `json:"user_info" orm:"rel(fk)"` // 发布用户的信息 ,根据这个查看发布用户的信息
	Expose         bool      `json:"expose"`                  // 是否是公开的
	ContentIllegal bool      `json:"content_illegal"`         // 内容是否符合
	CopyRight      bool      `json:"copy_right"`              // 版权是否正确
	PublishTime    time.Time `json:"publish_time"`            // 发布时间
	Link           string    `json:"link"`                    // 书籍的外连接
	Types          string    `json:"types"`                   // 数据的类型
	InspectTime    time.Time `json:"-"`                       // 审查的日期
	Reward         int       `json:"reward"`                  // 阅读完成的奖励
	Cost           int       `json:"cost"`                    // 阅读完成需要消费的资源
	Author         string    `json:"author"`                  // 作者的名字
	SaveName       string    `json:"-"`                       // 保存在磁盘的名字
	Del            bool      `json:"del"`                     // 该书是否已经被删除了
}

// 持久化到数据库
func (this *PublishBook) Insert() (id int64, err error) {
	return orm.NewOrm().Insert(this)
}
