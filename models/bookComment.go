package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

// 用户评价书籍的内容
type BookComment struct {
	Id          int64        `json:"id" orm:"pk"`     // 评论的id
	BookProfile *BookProfile `json:"-" orm:"rel(fk)"` // 评价书籍的id
	UserInfo    *UserInfo    `json:"-" orm:"rel(fk)"` // 用户的信息
	CommentTime time.Time                             // 评价的时间
	Content     string                                // 评价的内容
}

func (this *BookComment) Insert() (id int64, err error) {
	return orm.NewOrm().Insert(this)
}
