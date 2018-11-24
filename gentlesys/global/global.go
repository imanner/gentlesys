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
var OnePageElementCount int //每页展示条数量
var CachePagesNums int      //最多显示的导航索引条数
var FlushNumsLimit int      //首页刷新的值，当首页新增到FlushNumsLimit后，开始刷新
var AccessTimesLimit int    //首页缓存量超过改值后，开始刷新
var MaxNoticeShowNums int   //首页最多展示的公告数。
var IsNginxCache bool
var NginxAccessLogPath string //nginx的日志存放处
var NginxAccessFlushTimes int //nginx更新主页（清楚缓存）的时间周期

type RecordIndex struct {
	Ref      string
	Title    string
	IsActive string
}

func init() {
	OnePageElementCount = GetIntFromCfg("cache::OnePageElementCount", 50)
	CachePagesNums = GetIntFromCfg("cache::CachePagesNums", 10)
	FlushNumsLimit = GetIntFromCfg("cache::FlushNumsLimit", 30)
	AccessTimesLimit = GetIntFromCfg("cache::AccessTimesLimit", 100)
	MaxNoticeShowNums = GetIntFromCfg("cache::MaxNoticeShowNums", 5)
	isNgnix := GetIntFromCfg("common::cacheMode", 1)
	if isNgnix == 2 {
		//使用nginx缓存
		IsNginxCache = true
	}
	NginxAccessLogPath = GetStringFromCfg("nginx::NginxAccessLogPath", "")
	NginxAccessFlushTimes = GetIntFromCfg("nginx::NginxAccessFlushTimes", 30)
	//初步认为30分钟刷新一次比较好
	if NginxAccessFlushTimes >= 60 {
		NginxAccessFlushTimes = 30
	}

}

//将总条数转换为页数
func tranNums2Page(nums int) int {
	var totalPages int
	/*自己做一个整数的向上取整*/
	if nums%OnePageElementCount == 0 {
		totalPages = nums / OnePageElementCount
	} else {
		totalPages = nums/OnePageElementCount + 1
	}
	return totalPages
}

//CreateNavIndex函数的内部功能函数
func CreateNavIndexByPages(curPage int, totalPages int, urlPrex string, urlArgFiele string) (records []RecordIndex, prev string, next string) {

	recordIndexList := make([]RecordIndex, CachePagesNums+2)

	integerPart := curPage / CachePagesNums
	remainderPart := curPage % CachePagesNums

	//如果总数小于CachePagesNums页，不需要任何前后..
	if totalPages < CachePagesNums {
		for i, _ := range recordIndexList[:totalPages] {
			recordIndexList[i].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, i)
			recordIndexList[i].Title = fmt.Sprintf("%d", i)
		}
		recordIndexList[remainderPart].IsActive = "active"

		var prePage string
		if curPage <= 0 {
			prePage = "#没有了"
		} else if curPage == 1 {
			prePage = fmt.Sprintf("/%s", urlPrex)
		} else {
			prePage = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage-1)
		}

		var nextPage string
		if curPage >= (totalPages - 1) {
			nextPage = "#没有了"
		} else {
			nextPage = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage+1)
		}
		//第0页直接用主要代替
		recordIndexList[0].Ref = fmt.Sprintf("/%s", urlPrex)
		return recordIndexList[:totalPages], prePage, nextPage
	} else {
		//第1排只需要后面的..
		var nums int
		if integerPart == 0 {
			for i, _ := range recordIndexList[:CachePagesNums] {
				recordIndexList[i].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, i)
				recordIndexList[i].Title = fmt.Sprintf("%d", i)
				nums++
			}
			recordIndexList[remainderPart].IsActive = "active"
			recordIndexList[nums].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, nums)
			recordIndexList[nums].Title = ".."
			nums++

			var prePage string
			if curPage <= 0 {
				prePage = "#没有了"
			} else if curPage == 1 {
				prePage = fmt.Sprintf("/%s", urlPrex)
			} else {

				prePage = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage-1)
			}
			//第0页直接用主要代替
			recordIndexList[0].Ref = fmt.Sprintf("/%s", urlPrex)
			return recordIndexList[:nums], prePage, fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage+1)
		}

		//中间的有前后的..
		maxInterger := totalPages / CachePagesNums

		if integerPart > 0 && integerPart < maxInterger {
			recordIndexList[0].Title = ".."
			recordIndexList[0].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, integerPart*CachePagesNums-1)
			nums++
			for i, _ := range recordIndexList[1 : CachePagesNums+1] {
				recordIndexList[nums].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, i+integerPart*CachePagesNums)
				recordIndexList[nums].Title = fmt.Sprintf("%d", i+integerPart*CachePagesNums)
				nums++
			}
			recordIndexList[remainderPart+1].IsActive = "active"
			recordIndexList[nums].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, (integerPart+1)*CachePagesNums)
			recordIndexList[nums].Title = ".."

			return recordIndexList, fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage-1), fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage+1)
		}

		//最后一行只有前面的..
		if integerPart >= maxInterger {
			recordIndexList[0].Title = ".."
			recordIndexList[0].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, integerPart*CachePagesNums-1)
			nums++
			for i, _ := range recordIndexList[1 : CachePagesNums+1] {
				recordIndexList[nums].Ref = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, i+integerPart*CachePagesNums)
				recordIndexList[nums].Title = fmt.Sprintf("%d", i+integerPart*CachePagesNums)
				nums++
				//-1是为了跳过最前面的..计数
				if nums+integerPart*CachePagesNums-1 >= totalPages {
					break
				}
			}
			recordIndexList[remainderPart+1].IsActive = "active"

			var nextPage string
			if curPage >= (totalPages - 1) {
				nextPage = "#没有了"
			} else {
				nextPage = fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage+1)
			}

			return recordIndexList[:nums], fmt.Sprintf("/%s%s=%d", urlPrex, urlArgFiele, curPage-1), nextPage
		}
	}
	return nil, "", ""
}

//负责生成多页的分页栏显示内容,totalNums总记录数,isPageNums,urlPrex url的前缀，urlArgFiele 参数字段
func CreateNavIndexByNums(curPage int, totalNums int, urlPrex string, urlArgFiele string) (records []RecordIndex, prev string, next string) {
	totalPages := tranNums2Page(totalNums)
	//总共的页数是totalPages页
	return CreateNavIndexByPages(curPage, totalPages, urlPrex, urlArgFiele)
}
