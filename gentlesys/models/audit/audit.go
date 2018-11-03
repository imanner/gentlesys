package audit

import (
	"gentlesys/global"
	"strconv"
	"strings"
)

var ArticleDir string

//审计相关的功能在此
var adminMap map[int]int

//common配置表中整数的配置放这里
var cfgIntMap map[string]int
var cfgStrMap map[string]string

func GetCommonIntCfg(key string) int {
	return cfgIntMap[key]
}

//读取common中字符的配置
func GetCommonStrCfg(key string) string {
	return cfgStrMap[key]
}

func init() {
	//读取管理员列表
	adminMap = make(map[int]int)
	manager := global.GetStringFromCfg("common::managerlist", "")
	mList := strings.Split(manager, ",")

	for _, v := range mList {
		i, _ := strconv.Atoi(v)
		adminMap[i] = 1
	}

	cfgIntMap = make(map[string]int)
	cfgStrMap = make(map[string]string)

	//读取一个玩家每日最大发帖量
	aUserdayMax := global.GetIntFromCfg("common::aUserDayMaxArticle", 30)
	cfgIntMap["aUserDayMaxArticle"] = aUserdayMax

	ArticleDir = global.GetStringFromCfg("common::articleDirPath", "")
	cfgStrMap["articleDirPath"] = ArticleDir

}

func IsAdmin(id int) bool {
	//logs.Error("info", id, adminMap[int(id)])
	if adminMap[id] == 1 {
		return true
	}
	return false
}
