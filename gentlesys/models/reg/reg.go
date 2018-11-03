package reg

import (
	//	"fmt"
	"regexp"
	//	"github.com/astaxie/beego/logs"
)

//去掉kindeditor中一些一直的bug格式

//去掉font-family:&quot;
func ReplaceRegString(reg string, src string, dst string) string {
	regx := regexp.MustCompile(reg)
	return regx.ReplaceAllString(src, dst)
}

//去掉quot //去掉多余的空的<span></span>
func DelErrorString(text string) string {
	reg := `font-family:\s*&quot;`
	//logs.Error(text)

	ret := ReplaceRegString(reg, text, `font-family:Microsoft YaHei;`)

	reg = `<span[^<]*></span>`
	ret = ReplaceRegString(reg, ret, ``)

	//logs.Error(ret)
	return ret
}

//给图片加上动态适配屏幕
func AddImagAutoClass(text string) string {
	reg := `<img src=`
	//logs.Error(text)
	ret := ReplaceRegString(reg, text, `<img class="img-responsive center-block" src=`)
	//logs.Error(ret)
	return ret
}
