package main

import (
	_ "reading/routers"

	"github.com/astaxie/beego"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// 增加静态资源文件
	// url 外部访问路径
	// path 本地资源文件
	beego.SetStaticPath("/static","static")
	beego.Run()
}
