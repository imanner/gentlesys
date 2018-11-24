package navigation

//与导航相关的一切都放在这里。还包括页面最上面的一些软导航
import (
	"fmt"
	"gentlesys/global"

	"github.com/astaxie/beego/logs"
)

type navNode struct {
	Name string
	Href string
}

type navDataSt struct {
	NavHead  string
	NavNodes []navNode
}

var navCommon string

func init() {
	data := getNavFromCfg()
	navCommon = global.CreateNav(`views/tpl/nav.tpl`, `views/tpl/tmp/nav.txt`, data)
	mainNav = getMainPageNavCfg()
}

func GetNav() string {
	if navCommon == "" {
		data := getNavFromCfg()
		navCommon = global.CreateNav(`views/tpl/nav.tpl`, `views/tpl/tmp/nav.txt`, data)
	}
	return navCommon
}

func getNavFromCfg() *navDataSt {
	nums := global.GetIntFromCfg("nav::nums", 0)

	if nums > 0 {
		var data navDataSt

		data.NavHead = global.GetStringFromCfg("nav::navHead", "Gentlesys")

		data.NavNodes = make([]navNode, nums)

		for i := 0; i < nums; i++ {
			name_ := fmt.Sprintf("nav::id.%d.name", i)
			href_ := fmt.Sprintf("nav::id.%d.href", i)

			data.NavNodes[i].Name = global.GetStringFromCfg(name_, "Gentlesys")
			data.NavNodes[i].Href = global.GetStringFromCfg(href_, "#")

			if data.NavNodes[i].Name == "" || data.NavNodes[i].Href == "" {
				logs.Error("GetNavFromCfg error ...")
			}
		}
		return &data
	}
	return nil
}

var mainNav *[]navNode = nil

func GetMainPageNavData() *[]navNode {
	if mainNav == nil {
		mainNav = getMainPageNavCfg()
	}
	return mainNav
}

//获取首页导航的配置信息
func getMainPageNavCfg() *[]navNode {
	nums := global.GetIntFromCfg("pagenav::nums", 0)
	if nums > 0 {
		pageNavNodes := make([]navNode, nums)

		for i := 0; i < nums; i++ {
			name_ := fmt.Sprintf("pagenav::id.%d.name", i)
			href_ := fmt.Sprintf("pagenav::id.%d.href", i)

			pageNavNodes[i].Name = global.GetStringFromCfg(name_, "检查配置")
			pageNavNodes[i].Href = global.GetStringFromCfg(href_, "#")
		}
		return &pageNavNodes
	}
	return nil
}
