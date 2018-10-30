package main

import (
	_ "gentlesys/routers"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego"
)

func main() {

	//设置日志
	logs.SetLogger(logs.AdapterFile, `{"filename":"sys.log","level":4}`)
	logs.Async()

	beego.Run()
}
