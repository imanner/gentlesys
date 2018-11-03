package routers

import (
	"gentlesys/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})

	//有关主题的路由
	beego.Router("/subject:id:int", &controllers.SubjectController{})

	//发文章
	beego.Router("/article", &controllers.ArticleController{})

}
