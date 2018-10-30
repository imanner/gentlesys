package controllers

import (
	"gentlesys/global"
	"gentlesys/models/navigation"
	"gentlesys/subject"
	"strconv"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["Title"] = global.GetStringFromCfg("main::webname", "Gentlesys")
	c.Data["Navigation"] = navigation.GetNav()
	c.Data["Pagenav"] = navigation.GetMainPageNavData()
	c.Data["Subject"] = subject.GetMainPageSubjectData()
	c.TplName = "main.tpl"
}

//主题 /sjt:id?tid=xx
type SubjectController struct {
	beego.Controller
}

func (c *SubjectController) Get() {

	sid := c.Ctx.Input.Param(":id")

	//使用/sjt:id?tid=xx访问
	numId, err := strconv.Atoi(sid)
	if err != nil || numId >= subject.GetMaxSubjectId() {
		logs.Error(err, sid)
		c.Abort("401")
		return
	}

	subobj := subject.GetSubjectById(numId)

	c.Data["Title"] = subobj.Name
	c.Data["Navigation"] = navigation.GetNav()
	//c.Data["Pagenav"] = navigation.GetMainPageNavData()
	c.Data["HrefSub"] = subobj.Href
	c.Data["SubName"] = subobj.Name
	c.Data["Topic"] = subject.GetMainPageSubjectData()
	c.TplName = "subject.tpl"
}
