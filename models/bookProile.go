package models

import "github.com/astaxie/beego/orm"

// 书单中书籍的基本信息
type BookProfile struct {
	Id          int64        `json:"id" orm:"pk"`                // 书籍的id
	ReadingList *ReadingList `json:"reading_list" orm:"rel(fk)"` // 书单的id
	Name        string       `json:"name"`                       // 书名
	Author      string       `json:"author"`                     // 作者
	Instruction string       `json:"instruction"`                // 简介
	Link        string       `json:"link"`                       // 连接
	Types       string       `json:"types"`                      // 所属的类型
}


func (this *BookProfile ) Insert( ) ( id int64 , err error) {
	return orm.NewOrm().Insert(this)
}