package routers

import (
	"gentlesys/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})

	//有关主题的路由
	beego.Router("/sjt:id:int", &controllers.SubjectController{})
}
