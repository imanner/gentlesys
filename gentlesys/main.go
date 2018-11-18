package main

import (
	_ "gentlesys/routers"

	"gentlesys/store"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego"
)

func IsDeled(obj *store.CommentData) bool {
	if obj.IsDel != nil {
		return *(obj.IsDel)
	} else {
		return false
	}

}

func main() {

	//增加模板函数
	beego.AddFuncMap("isdeled", IsDeled)

	//设置日志
	logs.SetLogger(logs.AdapterFile, `{"filename":"sys.log","level":4}`)
	logs.Async()

	//开启连接管理
	beego.BConfig.WebConfig.Session.SessionOn = true

	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600 * 24 * 7  //一周
	beego.BConfig.WebConfig.Session.SessionCookieLifeTime = 3600 * 24 * 7 //
	beego.BConfig.WebConfig.Session.SessionName = "gentlesys"

	/*不允许自动设置cookie，自己控制session，不要总是第一次就设置session，而只是在登录时设置一次session*/
	beego.BConfig.WebConfig.Session.SessionAutoSetCookie = false

	beego.BConfig.WebConfig.Session.SessionProvider = "file"
	beego.BConfig.WebConfig.Session.SessionProviderConfig = "./sess"

	beego.Run()
}
