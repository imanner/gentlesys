package controllers

import (
	"fmt"
	"gentlesys/global"
	"gentlesys/models/audit"
	"gentlesys/models/navigation"
	"gentlesys/models/sqlsys"
	"gentlesys/subject"
	"strconv"

	"github.com/astaxie/beego/validation"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego"
)

//统一校验错误的接口
func DealParameterCheck(obj interface{}, errorStr string, c *beego.Controller) bool {
	valid := validation.Validation{}
	b, err := valid.Valid(obj)
	if err != nil {
		c.Ctx.WriteString(errorStr)
		logs.Error(errorStr)
		return false
	}

	if !b {
		for _, err := range valid.Errors {
			logs.Error(err.Key, err.Message)
		}
		c.Ctx.WriteString(errorStr)
		logs.Error(errorStr)
		return false
	}

	return true
}

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

//主题 subject:id?page=xx
type SubjectController struct {
	beego.Controller
}

func (c *SubjectController) Get() {

	sid := c.Ctx.Input.Param(":id")

	//或/subject:id?page=xx访问
	numId, err := strconv.Atoi(sid)
	if err != nil || !subject.IsSubjectIdExist(numId) {
		logs.Error(err, sid)
		c.Abort("401")
		return
	}

	//0表示回到首页
	pageIndex, _ := c.GetInt("page", 0)

	urlPrex := fmt.Sprintf("subject%s", sid)

	records, prev, next := global.CreateNavIndex(pageIndex, 100, urlPrex)
	if records != nil {
		c.Data["RecordIndexs"] = records
		c.Data["PrePage"] = prev
		c.Data["NextPage"] = next
	}

	subobj := subject.GetSubjectById(numId)

	c.Data["Title"] = subobj.Name
	c.Data["Navigation"] = navigation.GetNav()
	//c.Data["Pagenav"] = navigation.GetMainPageNavData()
	c.Data["Args"] = fmt.Sprintf("?sid=%s", sid)
	c.Data["HrefSub"] = subobj.Href
	c.Data["SubName"] = subobj.Name
	c.Data["Topic"] = subject.GetMainPageSubjectData()
	c.TplName = "subject.tpl"
}

type ArticleController struct {
	beego.Controller
}

//进入到写文章的界面
func (c *ArticleController) Get() {
	/*v := c.GetSession("user")
	if v == nil {
		c.Data["Nav"] = nav.CommolNav
		c.TplName = "auth.tpl"
	} else {*/
	//已经登录了，走到欢迎页面
	//if c.Ctx.GetCookie("user") == "" {
	//	c.Ctx.SetCookie("user", v.(string))
	//}

	sid, _ := c.GetInt("sid", -1)

	if sid == -1 || !subject.IsSubjectIdExist(sid) {
		logs.Error("ArticleController no sid exist", sid)
		c.Abort("401")
		return
	}

	c.Data["Nav"] = ""
	c.Data["UserId"] = 123
	c.Data["UserName"] = "123"
	c.Data["TopicType"] = subject.GetTopList()
	c.Data["Sid"] = sid

	c.TplName = "topic.tpl"
	//}
}

//发文章写数据库，然后将新生成的网页地址发给前端；用户提交的分享数据
func (c *ArticleController) Post() {

	u := &sqlsys.CommitArticle{}
	if err := c.ParseForm(u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {

		if !DealParameterCheck(u, "[3]数据格式异常，请修正！", &c.Controller) {
			return
		}
		//如果存在文章id，说明是修改，不是新增，走更新流程
		//暂时不写，在修改是加入

		if u.Story == "" {
			c.Ctx.WriteString("[3]博文数据格式异常，请检查故事文字长度，请修正！")
			logs.Error("[3]博文数据格式异常，请修正！")
			return
		}

		//新增流程
		var userAudit sqlsys.UserAudit

		userAudit.UserId = 1

		if !userAudit.ReadDb() {
			//没有该用户的审计记录，则插入一条记录
			userAudit.Insert()
		} else {
			//有记录，判断今天是否满足发布条件，否则不允许发布，防止数据库恶意写入。注意错误码[4]一般表示不能重试的那种错误，其他错误码随意。
			if !userAudit.IsAdmin() && userAudit.DayArticleNums > audit.GetCommonIntCfg("aUserDayMaxArticle") {
				c.Ctx.WriteString("[4]您今天发布的文章过多，为保证发布质量，请明天再来发布！")
				return
			}
		}

		u.UserId = userAudit.UserId
		r, _ := u.WriteDb()
		if r != 0 {
			//这里表示文章已经保存到数据库，原子更新数据库当前索引值
			//atomic.StoreUint32(&mysqlTool.ShareCureIndex, r)

			//更新用户的发帖记录
			userAudit.TlArticleNums++
			userAudit.DayArticleNums++

			userAudit.UpdataDayArticle()
			//将返回地址返回给客户端，让其跳转
			ret := fmt.Sprintf("[0]/aid%d", r)
			c.Ctx.WriteString(ret)

			//把主页main也刷新下，让用户能够实时看到自己的帖子
			//mysqlTool.UpdataMainPageDataRealTime(r, s)
			//处理匿名
			/*
				if s.Anonymity {
					s.ArName = "晒方网友"
				}
				cachemanager.ManagerCache.AddElementAtFront(s)
				//将返回地址返回给客户端，让其跳转,配合nginx清空缓存。放在RealTime里面去做
			*/
		} else {
			c.Ctx.WriteString("[1]保存数据库失败")
			logs.Error("[1]保存数据库失败")
		}
	}

}
