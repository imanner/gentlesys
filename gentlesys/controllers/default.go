package controllers

import (
	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	title, _ := beego.GetConfig("String", "main::webname", "Gentlesys")
	c.Data["Title"] = title
	c.TplName = "main.tpl"
}
