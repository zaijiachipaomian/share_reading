package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"reading/models"
	_ "reading/routers"

	"github.com/astaxie/beego"
)
// 初始化mysql 数据库
// 如果遇到错误,打印错误日志信息,退出
func initDataBaseMysql(){

	orm.Debug = true
	var err error
	err = orm.RegisterDriver("mysql", orm.DRMySQL)
	// 如果注册数据库启动失败,打印错误日志,退出程序
	if err != nil {
		fmt.Println("error orm.RegisterDriver ", err)
		os.Exit(2)
	}
	//err = 	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/park?charset=utf8&loc=Asia%2FShanghai", 30)
	// 获取数据库的地址.. 这里的地址从配置文件中读取
	dataSource := beego.AppConfig.String("source")
	// 注册 MySQL数据库, 使用的default别名
	err = orm.RegisterDataBase("default", "mysql", dataSource, 30)
	if err != nil {
		// 注册数据库失败, 打印错误信息退出程序
		fmt.Println("error orm.RegisterDataBase", err)
		os.Exit(2)
	}

	// 注册数据库的模型
	orm.RegisterModel(&models.UserInfo{}, &models.UploadBook{},&models.PublishBook{},&models.ReadingList{},&models.BookProfile{},&models.BookComment{})
}
func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// 注册数据库库
	initDataBaseMysql()
	// 增加静态资源文件
	// url 外部访问路径
	// path 本地资源文件
	beego.SetStaticPath("/static/*","static")
	beego.Run()
}
