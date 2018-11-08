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

	//注册
	beego.Router("/register", &controllers.RegisterController{})
	//登录
	beego.Router("/auth", &controllers.AuthController{})
	//退出
	beego.Router("/quit", &controllers.QuitController{})
	//找回密码
	beego.Router("/findpd", &controllers.FindPasswdController{})
	//请求重置密码
	beego.Router("/repasswd=:id:string", &controllers.RePasswdController{})
	//提交更新密码Post
	beego.Router("/updatepd", &controllers.UpdatePasswdController{})
}
