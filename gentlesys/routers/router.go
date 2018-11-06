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

	//浏览文章
	beego.Router("/browse", &controllers.BrowseController{})

	//评论文章
	beego.Router("/comment", &controllers.CommentController{})
}
