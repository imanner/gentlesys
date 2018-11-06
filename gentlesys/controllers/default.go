package controllers

import (
	"fmt"
	"gentlesys/cachemanager"
	"gentlesys/comment"
	"gentlesys/global"
	"gentlesys/models/audit"
	"gentlesys/models/navigation"
	"gentlesys/models/sqlsys"
	"gentlesys/subject"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

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

	records, prev, next := global.CreateNavIndex(pageIndex, subject.GetCurTotalTopicNums(numId), urlPrex)
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
	c.Data["Sid"] = subobj.UniqueId

	if pageIndex >= 0 && pageIndex < global.CachePagesNums {
		//如果是首页，首页特殊处理，因为首页可能实时发帖更新
		c.Data["Topic"] = cachemanager.CacheSubjectObjMaps[numId].ReadElementsWithPageNums(pageIndex)
	} else {
		//其他页呢，可以走ngnix的缓存页面，可以直接从数据库查询
		c.Data["Topic"] = (*sqlsys.Subject)(nil).GetTopicListPageNum(numId, pageIndex)
	}
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
	c.Data["TopicType"] = subject.GetTopicTyleList()
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
		r, topic := u.WriteDb()
		if r != 0 {
			//这里表示文章已经保存到数据库，原子更新数据库当前索引值
			subject.UpdateCurTopicIndex(u.SubId, r)

			//更新用户的发帖记录
			userAudit.TlArticleNums++
			userAudit.DayArticleNums++

			userAudit.UpdataDayArticle()
			//将返回地址返回给客户端，让其跳转
			ret := fmt.Sprintf("[0]/browse?sid=%d&aid=%d", u.SubId, r)
			c.Ctx.WriteString(ret)
			cachemanager.CacheSubjectObjMaps[u.SubId].AddElementAtFront(topic)
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

//浏览文章的路由
type BrowseController struct {
	beego.Controller
}

//获取评论
func (c *BrowseController) GetComment(filePath string) *[]*comment.CommentData {
	isExist := comment.CheckExists(filePath)
	if !isExist {
		return nil
	}

	fd, _ := os.OpenFile(filePath, os.O_RDONLY, 0644)
	defer fd.Close()

	ctobj := &comment.Comment{}
	ctobj.Fd = fd
	ret, _ := ctobj.GetOnePageComments(0)
	return ret
}

func (c *BrowseController) Get() {

	sid, _ := c.GetInt("sid", -1)
	aid, _ := c.GetInt("aid", -1)

	if sid == -1 || aid == -1 || !subject.IsSubjectIdExist(sid) {
		logs.Error("BrowseController err", sid, aid)
		c.Abort("401")
		return
	}

	ret, subobj := sqlsys.ReadSubjectFromDb(sid, aid)
	if 0 == ret {
		if subobj.Disable {
			c.Ctx.WriteString("[3]文章不符合审核规定，已经被禁用！")
			return
		}
		c.Data["Type"] = subobj.Type

		if subobj.Anonymity {
			c.Data["UserName"] = "匿名网友"
		} else {
			c.Data["UserName"] = subobj.UserName
		}

		c.Data["Title"] = subobj.Title
		c.Data["Nav"] = navigation.GetNav()
		c.Data["Date"] = subobj.Date

		subnodes := subject.GetSubjectById(sid)

		c.Data["HrefSub"] = subnodes.Href
		c.Data["SubName"] = subnodes.Name

		c.Data["Sid"] = sid
		c.Data["Aid"] = aid

		comments := c.GetComment(fmt.Sprintf("%s\\s%d_a%d", audit.GetCommonStrCfg("commentDirPath"), sid, aid))
		if comments != nil {
			c.Data["Comments"] = comments
		}

		if subobj.Path == "" {
			c.Data["Story"] = "很遗憾，用户没有留下TA的故事"
		} else {
			path := fmt.Sprintf("%s/%s", audit.ArticleDir, subobj.Path)
			if fileObj, err := os.Open(path); err == nil {
				defer fileObj.Close()
				if contents, err := ioutil.ReadAll(fileObj); err == nil {
					result := strings.Replace(string(contents), "\n", "", 1)
					c.Data["Story"] = result
				}

			} else {
				c.Data["Story"] = "很遗憾，用户没有留下TA的故事"
			}
		}
		c.TplName = "browse.tpl"

	} else {
		c.Abort("401")
	}
}

//评论，从客户端提交过来的数据
type Comment struct {
	ArtiId int    `form:"aid_" valid:"Required“`     //文章Id
	SubId  int    `form:"sid_" valid:"Required“`     //主题id
	Value  string `form:"comment_" valid:"Required“` //主题id
}

//评论文章的路由
type CommentController struct {
	beego.Controller
}

//提交评论文章
func (c *CommentController) Post() {

	u := &Comment{}
	if err := c.ParseForm(u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {

		if !DealParameterCheck(u, "[3]数据格式异常，请修正！", &c.Controller) {
			return
		}
	}

	aData := &comment.CommentData{}
	aData.Content = &u.Value
	filePath := fmt.Sprintf("%s\\s%d_a%d", audit.GetCommonStrCfg("commentDirPath"), u.SubId, u.ArtiId)

	ctobj := comment.GetCommentHandlerByPath(fmt.Sprintf("%s_%s", u.SubId, u.ArtiId))

	ctobj.Mutex.Lock()
	defer ctobj.Mutex.Unlock()

	isExist := comment.CheckExists(filePath)

	fd, _ := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	defer fd.Close()

	ctobj.Fd = fd
	if !isExist {
		ctobj.InitMcData()
	}
	if ctobj.AddOneComment(aData) {
		c.Ctx.WriteString("[0]提交点评成功")
	} else {
		c.Ctx.WriteString("[1]提交点评失败")
	}
}
