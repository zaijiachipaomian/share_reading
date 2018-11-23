package controllers

import (
	"github.com/astaxie/beego"
	"reading/models"
)

// 响应的参数
var (
	errorRes400 = models.ResponseMessage{Detail: "参数错误,请重新提交", Code: 400}
	errorRes401 = models.ResponseMessage{Detail: "请登录", Code: 401}
	errorRes403 = models.ResponseMessage{Detail: "本次访问已经被积极拒绝", Code: 403}
	errorRes404 = models.ResponseMessage{Detail: "访问路径出错", Code: 404}
	// 内部错误
	errorRes500 = models.ResponseMessage{Detail: "未知错误", Code: 500}
	errorRes503 = models.ResponseMessage{Detail: "未知错误", Code: 503}
)

type ErrorController struct {
	beego.Controller
}

// 请求参数错误
// 包括用户提交的数据格式
// 参数缺失
func (this *ErrorController) Error400() {

	response(&this.Controller, errorRes400)
	return
}

// 需要验证 auth
func (this *ErrorController) Error401() {

	response(&this.Controller, errorRes401)
	return
}

// 禁止访问
// 无权限访问本路径
func (this *ErrorController) Error403() {
	response(&this.Controller, errorRes403)
	return
}

// 找不到路径
func (this *ErrorController) Error404() {
	response(&this.Controller, errorRes404)
	return
}

// 500 内部错误
func (this *ErrorController) Error500() {
	response(&this.Controller, errorRes500)
	return
}

// 503 内部错误
func (this *ErrorController) Error503() {
	response(&this.Controller, errorRes503)
	return
}

// 相应 处理
// 采用json 格式 返回 数据
func response(controller *beego.Controller, data interface{}) {
	controller.Data["json"] = data
	controller.ServeJSON(true)
}
