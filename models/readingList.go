package models

import "github.com/astaxie/beego/orm"

type ReadingList struct {
	Id          int64          `json:"id" orm:"pk"`
	UserInfo    *UserInfo      `json:"-" orm:"rel(fk)"`
	Name        string         `json:"name"`
	Instruction string         `json:"instruction"`
	Types       string         `json:"types"`
	BookProfile []*BookProfile `json:"book_profile" orm:"reverse(many);"`
	Expose      bool           `json:"expose"` // 该书单是否是公开的
}

func (this *ReadingList) Insert() (id int64, err error) {
	return orm.NewOrm().Insert(this)
}
