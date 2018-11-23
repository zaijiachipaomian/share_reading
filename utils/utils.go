package utils

import (
	"encoding/json"
	"errors"
	"regexp"
)

const (
	PhonePattern = `^(1[3|4|5|8|6][0-9]\d{8,8})$`
)

// 将一个对象序列化成一个字符串
func Marshal2JSONString(obj interface{}) (str string, err error) {
	data, err := json.Marshal(obj)
	// 如果遇到错误,返回空串,和错误
	if err != nil {
		return str, err
	}
	str = string(data)
	// 返回序列化的字符串,和空的错误
	return str, err
}

// 检验手机号码是否合格
// phone 是手机号码
// pattern 是正则表达式的模式
// 返回值
// ok 如果为true 表示手机号码符合正则表达式的校验, 如果为false 表示该手机号码不通过这则表达式的检验
// err 如果遇到错误,则返回error
func RegexpValidPhone(phone string, pattern string) (ok bool, err error) {
	if len(phone) != 11 {
		return false, errors.New("phone number is error ")
	}
	ok, err = regexp.MatchString(pattern, phone)
	return ok, err
}
