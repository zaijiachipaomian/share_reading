package models

// 响应的消息
// Detail 使用interface{} 类型适用于任何类型作为值
// Code   响应的状态码
type ResponseMessage struct {
	Detail interface{} `json:"detail"`
	Code   int         `json:"code"`
}
