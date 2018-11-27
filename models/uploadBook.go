package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

//ID, 用户的ID, 书名, 上传日期, 大小, 保存的连接名(时间戳+用户id格式化+.pdf)

type UploadBook struct {
	Id         int64     `json:"id" orm:"pk,column(id)"`
	UserInfo   *UserInfo `orm:"rel(fk)"`
	BookName   string    `json:"book_name"`
	UploadTime time.Time `json:"upload_time"`
	Size       int64     `json:"size"`
	SaveName   string    `json:"-"`
}


func (this *UploadBook) Insert() ( id int64 , err error){
	return orm.NewOrm().Insert(this )
}