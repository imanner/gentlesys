package global

import (
	"io/ioutil"
	"os"
	"text/template"

	"github.com/astaxie/beego"
)

//这个文件存放全局都要使用的函数

func GetIntFromCfg(key string, deft int) int {
	v, _ := beego.GetConfig("Int", key, deft)
	return v.(int)
}

func GetStringFromCfg(key string, deft string) string {
	v, _ := beego.GetConfig("String", key, deft)
	return v.(string)
}

//用于处理模板的函数
func CreateNav(tplName string, saveName string, tplData interface{}) string {

	var err error
	var t *template.Template
	t, err = template.ParseFiles(tplName) //从文件创建一个模板
	if err != nil {
		panic(err)
	}

	if fileObj, err1 := os.Create(saveName); err1 == nil {
		defer fileObj.Close()
		err = t.Execute(fileObj, tplData)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err1)
	}

	ret, _ := ioutil.ReadFile(saveName)
	return string(ret)
}
