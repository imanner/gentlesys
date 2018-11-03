package global

import (
	"fmt"
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

//专门负责生成页面index的函数
const OnePageShareCount = 50 //每页展示50条
const NavIndexShowNums = 10  //最多显示的导航索引条数

type RecordIndex struct {
	Ref      string
	Title    string
	IsActive string
}

//负责生成多页的分页栏显示内容
func CreateNavIndex(curPage int, totalNums int, urlPrex string) (records []RecordIndex, prev string, next string) {
	var totalPages int
	/*自己做一个整数的向上取整*/
	if totalNums%OnePageShareCount == 0 {
		totalPages = totalNums / OnePageShareCount
	} else {
		totalPages = totalNums/OnePageShareCount + 1
	}
	//总共的页数是totalPages页

	//前后分别有一个..导航
	recordIndexList := make([]RecordIndex, NavIndexShowNums+2)

	integerPart := curPage / NavIndexShowNums
	remainderPart := curPage % NavIndexShowNums

	//如果总数小于10页，不需要任何前后..
	if totalPages < NavIndexShowNums {
		for i, _ := range recordIndexList[:totalPages] {
			recordIndexList[i].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, i)
			recordIndexList[i].Title = fmt.Sprintf("%d", i)
		}
		recordIndexList[remainderPart].IsActive = "active"

		var prePage string
		if curPage <= 0 {
			prePage = "#没有了"
		} else if curPage == 1 {
			prePage = fmt.Sprintf("/%s", urlPrex)
		} else {
			prePage = fmt.Sprintf("/%s?page=%d", urlPrex, curPage-1)
		}

		var nextPage string
		if curPage >= (totalPages - 1) {
			nextPage = "#没有了"
		} else {
			nextPage = fmt.Sprintf("/%s?page=%d", urlPrex, curPage+1)
		}
		//第0页直接用主要代替
		recordIndexList[0].Ref = fmt.Sprintf("/%s", urlPrex)
		return recordIndexList[:totalPages], prePage, nextPage
	} else {
		//第1排只需要后面的..
		var nums int
		if integerPart == 0 {
			for i, _ := range recordIndexList[:NavIndexShowNums] {
				recordIndexList[i].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, i)
				recordIndexList[i].Title = fmt.Sprintf("%d", i)
				nums++
			}
			recordIndexList[remainderPart].IsActive = "active"
			recordIndexList[nums].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, nums)
			recordIndexList[nums].Title = ".."
			nums++

			var prePage string
			if curPage <= 0 {
				prePage = "#没有了"
			} else if curPage == 1 {
				prePage = fmt.Sprintf("/%s", urlPrex)
			} else {

				prePage = fmt.Sprintf("/%s?page=%d", urlPrex, curPage-1)
			}
			//第0页直接用主要代替
			recordIndexList[0].Ref = fmt.Sprintf("/%s", urlPrex)
			return recordIndexList[:nums], prePage, fmt.Sprintf("/%s?page=%d", urlPrex, curPage+1)
		}

		//中间的有前后的..
		maxInterger := totalPages / NavIndexShowNums

		if integerPart > 0 && integerPart < maxInterger {
			recordIndexList[0].Title = ".."
			recordIndexList[0].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, integerPart*NavIndexShowNums-1)
			nums++
			for i, _ := range recordIndexList[1 : NavIndexShowNums+1] {
				recordIndexList[nums].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, i+integerPart*NavIndexShowNums)
				recordIndexList[nums].Title = fmt.Sprintf("%d", i+integerPart*NavIndexShowNums)
				nums++
			}
			recordIndexList[remainderPart+1].IsActive = "active"
			recordIndexList[nums].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, (integerPart+1)*NavIndexShowNums)
			recordIndexList[nums].Title = ".."

			return recordIndexList, fmt.Sprintf("/%s?page=%d", urlPrex, curPage-1), fmt.Sprintf("/%s?page=%d", urlPrex, curPage+1)
		}

		//最后一行只有前面的..
		if integerPart >= maxInterger {
			recordIndexList[0].Title = ".."
			recordIndexList[0].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, integerPart*NavIndexShowNums-1)
			nums++
			for i, _ := range recordIndexList[1 : NavIndexShowNums+1] {
				recordIndexList[nums].Ref = fmt.Sprintf("/%s?page=%d", urlPrex, i+integerPart*NavIndexShowNums)
				recordIndexList[nums].Title = fmt.Sprintf("%d", i+integerPart*NavIndexShowNums)
				nums++
				//-1是为了跳过最前面的..计数
				if nums+integerPart*NavIndexShowNums-1 >= totalPages {
					break
				}
			}
			recordIndexList[remainderPart+1].IsActive = "active"

			var nextPage string
			if curPage >= (totalPages - 1) {
				nextPage = "#没有了"
			} else {
				nextPage = fmt.Sprintf("/?page=%d", curPage+1)
			}

			return recordIndexList[:nums], fmt.Sprintf("/%s?page=%d", urlPrex, curPage-1), nextPage
		}
	}

	return nil, "", ""
}
