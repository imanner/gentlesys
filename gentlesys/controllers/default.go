package controllers

import (
	"fmt"
	"gentlesys/cachemanager"
	"gentlesys/comment"
	"gentlesys/global"
	"gentlesys/models/audit"
	"gentlesys/models/mail"
	"gentlesys/models/navigation"
	"gentlesys/models/reg"
	"gentlesys/models/sqlsys"
	"gentlesys/subject"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

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
	c.Data["Title"] = "Gentlesys"
	c.Data["Navigation"] = navigation.GetNav()
	pn := navigation.GetMainPageNavData()
	if pn != nil {
		c.Data["Pagenav"] = pn
	}
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

	records, prev, next := global.CreateNavIndexByNums(pageIndex, subject.GetCurTotalTopicNums(numId), urlPrex, "?page")
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
		topics := cachemanager.CacheSubjectObjMaps[numId].ReadElementsWithPageNums(pageIndex)
		if len(topics) == 0 {
			c.Data["NoMore"] = true
		} else {
			c.Data["Topic"] = topics
		}
	} else {
		//其他页呢，可以走ngnix的缓存页面，可以直接从数据库查询
		topics := (*sqlsys.Subject)(nil).GetTopicListPageNum(numId, pageIndex)
		if len(*topics) == 0 {
			c.Data["NoMore"] = true
		} else {
			c.Data["Topic"] = topics
		}
	}
	c.TplName = "subject.tpl"
}

type ArticleController struct {
	beego.Controller
}

//进入到写文章的界面
func (c *ArticleController) Get() {
	v := c.GetSession("user")
	if v == nil {
		c.Data["Navigation"] = navigation.GetNav()
		//没有登录，先跳转到登录
		c.TplName = "auth.tpl"
		return
	}

	sid, _ := c.GetInt("sid", -1)

	if sid == -1 || !subject.IsSubjectIdExist(sid) {
		logs.Error("ArticleController no sid exist", sid)
		c.Abort("401")
		return
	}

	c.Data["Navigation"] = navigation.GetNav()
	id := c.GetSession("id")
	c.Data["UserId"] = id.(int)
	c.Data["UserName"] = v.(string)
	c.Data["TopicType"] = subject.GetTopicTyleList()
	c.Data["Sid"] = sid

	c.TplName = "topic.tpl"
}

//发文章写数据库，然后将新生成的网页地址发给前端；用户提交的分享数据
func (c *ArticleController) Post() {
	v := c.GetSession("id")
	if v == nil {
		c.Data["Navigation"] = navigation.GetNav()
		//没有登录，先跳转到登录
		c.TplName = "auth.tpl"
		return
	}

	u := &sqlsys.CommitArticle{}
	if err := c.ParseForm(u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {

		if !DealParameterCheck(u, "[3]数据格式异常，请修正！", &c.Controller) {
			return
		}
		//如果存在文章id，说明是修改，不是新增，走更新流程
		//更新流程
		if u.ArtiId > 0 {

			//避免伪造作者改文章
			if v.(int) != u.UserId {
				ret := fmt.Sprintf("[3]只能编辑自己的帖子，不可伪造用户信息，伪造用户id是%d %d", v.(int), u.UserId)
				logs.Error(ret)
				c.Ctx.WriteString("[3]只能编辑自己的帖子，不可伪造用户信息")
				return
			}

			r := u.UpdateDb()
			if r {
				//将返回地址返回给客户端，让其跳转,配合nginx清空缓存
				//clearcache.ClearPath(fmt.Sprintf("/cure%d", u.ShareId))

				ret := fmt.Sprintf("[0]/browse?sid=%d&aid=%d", u.SubId, u.ArtiId)
				c.Ctx.WriteString(ret)
			} else {
				ret := fmt.Sprintf("[%d]保存数据库失败", r)
				c.Ctx.WriteString(ret)
				logs.Error(ret)
			}
			return
		}

		if u.Story == "" {
			c.Ctx.WriteString("[3]博文数据格式异常，请检查故事文字长度，请修正！")
			logs.Error("[3]博文数据格式异常，请修正！")
			return
		}

		//新增流程
		var userAudit sqlsys.UserAudit

		userAudit.UserId = v.(int)

		if !userAudit.ReadDb() {
			//没有该用户的审计记录，则插入一条记录
			userAudit.Insert()
		} else {
			if userAudit.Could {
				c.Ctx.WriteString("[4]你已经被禁言，不能再发帖！")
				return
			}
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
			ctobj := &comment.Comment{}
			aTopic := &comment.UserTopicData{}
			aTopic.Aid = proto.Int(r)
			aTopic.Sid = proto.Int(u.SubId)
			aTopic.Title = &topic.Title
			aTopic.Time = proto.String(time.Now().Format("2006-01-02 15:04:05"))

			filePath := fmt.Sprintf("%s\\u_%d", audit.GetCommonStrCfg("userTopicDirPath"), u.UserId)
			isExist := comment.CheckExists(filePath)

			fd, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
			if err == nil {
				defer fd.Close()
			} else {
				logs.Error("用户发帖列表保存失败")
				return //暂时不干啥
			}

			ctobj.Fd = fd
			if !isExist {
				ctobj.InitMcData()
			}
			if ok, _ := ctobj.AddOneUserTopic(aTopic); !ok {
				logs.Error("增加用户发帖列表保存失败")
			}
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

//获取评论,这个里面的异步读，其他地方可能异步写，要小心
func (c *BrowseController) GetComment(filePath string, pages int, sid int, aid int) *[]*comment.CommentData {
	isExist := comment.CheckExists(filePath)
	if !isExist {
		return nil
	}

	key := fmt.Sprintf("%s_%s", sid, aid)
	ctobj := comment.GetCommentHandlerByPath(key)
	defer comment.DelCommentHandlerByPath(key)
	//上下两个defer的位置顺序值得思考。读加读锁
	ctobj.Mutex.RLock()
	defer ctobj.Mutex.RUnlock()

	fd, _ := os.OpenFile(filePath, os.O_RDONLY, 0644)
	defer fd.Close()

	ctobj.Fd = fd
	ret, _ := ctobj.GetOnePageComments(pages)
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
		c.Data["Type"] = subject.GetTopicTyleById(subobj.Type)

		if subobj.Anonymity {
			c.Data["UserName"] = "匿名网友"
		} else {
			c.Data["UserName"] = subobj.UserName
		}

		c.Data["Title"] = subobj.Title
		c.Data["Navigation"] = navigation.GetNav()
		c.Data["Date"] = subobj.Date

		subnodes := subject.GetSubjectById(sid)

		c.Data["HrefSub"] = subnodes.Href
		c.Data["SubName"] = subnodes.Name

		c.Data["Sid"] = sid
		c.Data["Aid"] = aid
		c.Data["Args"] = fmt.Sprintf("?sid=%d", sid)

		//0表示回到首页
		pageIndex, _ := c.GetInt("page", 0)

		urlPrex := fmt.Sprintf("browse?sid=%d&aid=%d", sid, aid)

		commentFilePath := fmt.Sprintf("%s\\s%d_a%d", audit.GetCommonStrCfg("commentDirPath"), sid, aid)
		curCommentPageNums := comment.GetCommentNums(commentFilePath)
		//如果请求页超过最大评论页，则返回评论最后一页
		if pageIndex > (curCommentPageNums - 1) {
			pageIndex = curCommentPageNums - 1
		}
		if pageIndex < 0 {
			pageIndex = 0
		}

		records, prev, next := global.CreateNavIndexByPages(pageIndex, curCommentPageNums, urlPrex, "&page")
		if records != nil {
			c.Data["RecordIndexs"] = records
			c.Data["PrePage"] = prev
			c.Data["NextPage"] = next
		}
		//评论超过MaxMetaMcSize页，不能再留言。目前是20*512条
		if curCommentPageNums >= comment.MaxMetaMcSize {
			c.Data["CanReplay"] = false
		} else {
			c.Data["CanReplay"] = true
		}
		//获取评论
		comments := c.GetComment(commentFilePath, pageIndex, sid, aid)
		if comments != nil && len(*comments) > 0 {
			c.Data["Comments"] = comments
			c.Data["NoMore"] = false
		} else {
			c.Data["NoMore"] = true
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

		//如果是实时更新访问量，则更新
		if 1 == audit.GetCommonIntCfg("topicReadStatics") {
			subobj.ReadTimes++
			sqlsys.UpdateTopicReadStatics(sid, aid, subobj.ReadTimes, subobj.ReplyTimes)
		}

	} else {
		c.Abort("401")
	}
}

//评论，从客户端提交过来的数据
type Comment struct {
	ArtiId int    `form:"aid_" valid:"Required“`                              //文章Id
	SubId  int    `form:"sid_" valid:"Required“`                              //主题id
	Value  string `form:"comment_" valid:"Required;MinSize(1);MaxSize(1000)"` //评论内容
}

//评论文章的路由
type CommentController struct {
	beego.Controller
}

//提交评论
func (c *CommentController) Post() {
	v := c.GetSession("user")
	if v == nil {
		c.Ctx.WriteString("[4]你还没有登录，不能留言,请先登录...")
		return
	}
	u := &Comment{}
	if err := c.ParseForm(u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {

		if !DealParameterCheck(u, "[3]数据格式异常，请修正！", &c.Controller) {
			return
		}
	}

	//用户回复审计
	var userAudit sqlsys.UserAudit

	id := c.GetSession("id")
	userAudit.UserId = id.(int)

	if !userAudit.ReadDb() {
		//没有该用户的审计记录，则插入一条记录
		userAudit.Insert()
	} else {
		//用户被禁用
		if userAudit.Could {
			c.Ctx.WriteString("[4]你已经被禁言，不能再发帖！")
			return
		}
		//有记录，判断今天是否满足发布条件，否则不允许发布，防止数据库恶意写入。注意错误码[4]一般表示不能重试的那种错误，其他错误码随意。
		if !userAudit.IsAdmin() && userAudit.DayCommentTimes > audit.GetCommonIntCfg("aUserDayMaxComment") {
			c.Ctx.WriteString("[4]您今天发布的评论过多，为保证评论质量，请明天再来发布！")
			return
		}
	}

	aData := &comment.CommentData{}
	//去掉kindeditor非法的字符
	u.Value = reg.DelErrorString(u.Value)
	//图片加上自动适配
	u.Value = reg.AddImagAutoClass(u.Value)
	aData.Content = &u.Value
	aData.Time = proto.String(time.Now().Format("2006-01-02 15:04:05"))
	aData.UserName = proto.String(v.(string))
	filePath := fmt.Sprintf("%s\\s%d_a%d", audit.GetCommonStrCfg("commentDirPath"), u.SubId, u.ArtiId)

	key := fmt.Sprintf("%s_%s", u.SubId, u.ArtiId)
	ctobj := comment.GetCommentHandlerByPath(key)
	defer comment.DelCommentHandlerByPath(key)
	//上下两个defer的位置顺序值得思考，写加写锁
	ctobj.Mutex.Lock()
	defer ctobj.Mutex.Unlock()

	isExist := comment.CheckExists(filePath)

	fd, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err == nil {
		defer fd.Close()
	} else {
		c.Ctx.WriteString("[1]提交点评失败")
		return
	}

	ctobj.Fd = fd
	if !isExist {
		ctobj.InitMcData()
	}
	if ok, pages := ctobj.AddOneComment(aData); ok {
		//跳转到点评页面的最后一页，让用户看到自己的点评
		c.Ctx.WriteString(fmt.Sprintf("[0]/browse?sid=%d&aid=%d&page=%d", u.SubId, u.ArtiId, pages))

		userAudit.TlCommentTimes++
		userAudit.DayCommentTimes++

		userAudit.UpdataDayCommentTimes()

	} else {
		c.Ctx.WriteString("[1]提交点评失败")
	}
}

type AuthController struct {
	beego.Controller
}

//登录页面
func (c *AuthController) Get() {
	v := c.GetSession("user")
	if v == nil {
		c.Data["Navigation"] = navigation.GetNav()
		c.TplName = "auth.tpl"
	} else {
		//已经登录了，走到主页页面
		userName := c.Ctx.GetCookie("user")

		if userName == "" || userName == "游客" {
			c.Ctx.SetCookie("user", v.(string), beego.BConfig.WebConfig.Session.SessionCookieLifeTime)
		}

		c.Data["Title"] = "用户登录"
		c.Data["Navigation"] = navigation.GetNav()
		c.Data["Pagenav"] = navigation.GetMainPageNavData()
		c.Data["Subject"] = subject.GetMainPageSubjectData()
		c.TplName = "main.tpl"

	}

}

//请求登录的流程
func (c *AuthController) Post() {
	u := sqlsys.User{}
	if err := c.ParseForm(&u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {
		if !DealParameterCheck(u, "[3]账号或密码格式异常，请修正", &c.Controller) {
			return
		}

		v := c.GetSession("user")
		if v == nil {
			//第一次验证用户名与密码
			ret := u.CheckUserAuth()
			switch ret {
			case 0:
				//验证通过
				c.SetSession("id", u.Id)
				c.SetSession("user", u.Name)
				c.Ctx.SetCookie("user", u.Name, beego.BConfig.WebConfig.Session.SessionCookieLifeTime)
				//如果不设置项SessionName，则beego的session不会生效。因为我们的配置是SessionAutoSetCookie=false
				c.Ctx.SetCookie(beego.BConfig.WebConfig.Session.SessionName, c.CruSession.SessionID(), beego.BConfig.WebConfig.Session.SessionCookieLifeTime)
				c.Ctx.WriteString("[0]登录成功！")
			case sqlsys.ERR_NO_USERNAME:
				c.Ctx.WriteString("[1]登录错误: 账号或密码格式异常！")
			case sqlsys.ERR_AUTH_FAIL:
				//密码错误,注意此时u.Id是有值的
				u.Fail++
				u.UpdateFail()
				c.Ctx.WriteString(fmt.Sprintf("[2]登录错误: 密码错误%d次", u.Fail))
			case sqlsys.ERR_FAIL_FORBID:
				c.Ctx.WriteString("[4]登录错误: 失败次数过多，账号暂时被禁用，今天不能登陆！")
			default:
				c.Ctx.WriteString("[5]登录错误: 账号或密码错误！")
			}

		} else {
			c.Ctx.WriteString("[0]欢迎回来" + v.(string))
		}
	}
}

type RegisterController struct {
	beego.Controller
}

func (c *RegisterController) Get() {
	c.Data["Navigation"] = navigation.GetNav()
	c.Data["Title"] = "用户注册"
	c.TplName = "register.tpl"
}

func (c *RegisterController) Post() {
	u := sqlsys.User{}
	if err := c.ParseForm(&u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {
		//beego.Informational(u)
		if !DealParameterCheck(u, "[3]数据格式异常，请修正！", &c.Controller) {
			return
		}

		if u.CheckUserExist() {
			c.Ctx.WriteString("[1]账号名称已经被注册，请重新换一个")
			return
		}

		r := u.WriteDb()
		if r != 0 {

			c.SetSession("user", u.Name)
			c.SetSession("id", u.Id)
			c.Ctx.SetCookie("user", u.Name, beego.BConfig.WebConfig.Session.SessionCookieLifeTime)
			c.Ctx.SetCookie(beego.BConfig.WebConfig.Session.SessionName, c.CruSession.SessionID(), beego.BConfig.WebConfig.Session.SessionCookieLifeTime)

			c.Ctx.WriteString("[0]注册成功")
		} else {
			c.Ctx.WriteString("[1]注册失败")
			logs.Error("[1]保存数据库失败")
		}

	}
}

//退出登录
type QuitController struct {
	beego.Controller
}

func (c *QuitController) Get() {
	v := c.GetSession("user")
	if v != nil {
		//已经登录了，退出删除Session
		c.DestroySession()
		c.Ctx.SetCookie("user", "游客")
	}

	c.Data["Title"] = "用户登录"
	c.Data["Navigation"] = navigation.GetNav()
	c.Data["Pagenav"] = navigation.GetMainPageNavData()
	c.Data["Subject"] = subject.GetMainPageSubjectData()
	c.TplName = "main.tpl"
}

//找回密码的页面
type FindPasswdController struct {
	beego.Controller
}

func (c *FindPasswdController) Get() {
	c.Data["Navigation"] = navigation.GetNav()
	c.TplName = "passwd.tpl"
}

//找回密码时用户从客户端回传的结构体
type findPasswd struct {
	Name string `form:"name_" valid:"Required;MinSize(1);MaxSize(32)“`
}

//找回密码post页面，发送邮件到邮箱
func (c *FindPasswdController) Post() {
	u := findPasswd{}
	if err := c.ParseForm(&u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {
		if !DealParameterCheck(u, "[3]账号格式异常，请修正！", &c.Controller) {
			return
		}

		var aUser sqlsys.User
		aUser.Name = u.Name

		if 0 != aUser.FindMailByName() {
			c.Ctx.WriteString("[1]错误的账户名，请修正")
			logs.Error("[1]找回密码错误不存在该账户名")
		} else {
			if aUser.Mail == "" {
				c.Ctx.WriteString(fmt.Sprintf("[1]该用户注册时没有留下邮箱，无法找回密码"))
				return
			}
			var aPassinfo sqlsys.PasswdReset
			aPassinfo.Name = u.Name

			if aPassinfo.InsertByName() {
				data := fmt.Sprintf("访问网址<a href=\"%s\">%s/repasswd=%s</a>修改密码", aPassinfo.UserId, mail.WebDomainName, aPassinfo.UserId)
				if mail.SendMail(aUser.Mail, "Gentlesys 找回密码", data) {
					c.Ctx.WriteString(fmt.Sprintf("[0]重置连接已发送到邮箱地址：%s, 请尽快查收", aUser.Mail))

				} else {
					c.Ctx.WriteString(fmt.Sprintf("[1]发送找回密码邮件失败，可能没有该用户"))
				}
			}
		}
	}
}

//在重置页面中重置密码
type RePasswdController struct {
	beego.Controller
}

//重置密码的Url中必须要带一个 md5后的路径，这个是在数据库中的
func (c *RePasswdController) Get() {

	index := c.Ctx.Input.Param(":id")

	var aRePass sqlsys.PasswdReset
	aRePass.UserId = index
	//如果我们的数据库中没有这个id，说明是伪造的修改密码
	if aRePass.ReadDb() {
		c.Data["User"] = aRePass.Name
		c.Data["Id"] = aRePass.UserId
		c.Data["Navigation"] = navigation.GetNav()
		c.TplName = "repasswd.tpl"
	} else {
		c.Abort("401")
	}
}

type RePasswdInfo struct {
	Id     string `form:"id_" valid:"Required“`
	Passwd string `form:"passwd_" valid:"Required;MinSize(1);MaxSize(32)" `
}

//实质更新用户密码的处理
type UpdatePasswdController struct {
	beego.Controller
}

func (c *UpdatePasswdController) Post() {
	u := RePasswdInfo{}
	if err := c.ParseForm(&u); err != nil {
		c.Ctx.WriteString("[2]格式不对，请修正！")
	} else {
		if !DealParameterCheck(u, "[3]密码格式异常，请修正！", &c.Controller) {
			return
		}

		//先从id修改库中找到对应的用户的名称
		var aPasswdReset sqlsys.PasswdReset
		aPasswdReset.UserId = u.Id
		if aPasswdReset.ReadDb() {
			var aUser sqlsys.User
			aUser.Name = aPasswdReset.Name
			aUser.Passwd = u.Passwd
			//再根据用户名称去更新他的新密码
			if 0 == aUser.UpdatePasswdByName() {

				c.SetSession("id", u.Id)
				c.SetSession("user", aUser.Name)
				c.Ctx.SetCookie("user", aUser.Name, beego.BConfig.WebConfig.Session.SessionCookieLifeTime)
				c.Ctx.SetCookie(beego.BConfig.WebConfig.Session.SessionName, c.CruSession.SessionID(), beego.BConfig.WebConfig.Session.SessionCookieLifeTime)
				c.Ctx.WriteString("[0]更新密码成功")
				aPasswdReset.Delete()
			}

		} else {
			c.Ctx.WriteString("[1]更新密码失败")
		}

	}
}

//用户中心
type UserInfoController struct {
	beego.Controller
}

//获取评论,这个里面的异步读，其他地方可能异步写，要小心
func (c *UserInfoController) GetTopics(filePath string, pages int) *[]*comment.UserTopicData {
	isExist := comment.CheckExists(filePath)
	if !isExist {
		return nil
	}

	ctobj := &comment.Comment{}

	fd, _ := os.OpenFile(filePath, os.O_RDONLY, 0644)
	defer fd.Close()

	ctobj.Fd = fd
	ret, _ := ctobj.GetOnePageTopics(pages)
	return ret
}

func (c *UserInfoController) Get() {

	c.Data["Navigation"] = navigation.GetNav()
	v := c.GetSession("id")
	if v == nil {
		//没有登录，先跳转到登录
		c.TplName = "auth.tpl"
		return
	}

	//0表示回到首页
	pageIndex, _ := c.GetInt("page", 0)

	topicFilePath := fmt.Sprintf("%s\\u_%d", audit.GetCommonStrCfg("userTopicDirPath"), v.(int))
	curTopicPageNums := comment.GetCommentNums(topicFilePath)
	//如果请求页超过最大帖子页，则返回最后一页
	if pageIndex > (curTopicPageNums - 1) {
		pageIndex = curTopicPageNums - 1
	}
	if pageIndex < 0 {
		pageIndex = 0
	}

	records, prev, next := global.CreateNavIndexByPages(pageIndex, curTopicPageNums, "usif", "?page")
	if records != nil {
		c.Data["RecordIndexs"] = records
		c.Data["PrePage"] = prev
		c.Data["NextPage"] = next
	}

	topics := c.GetTopics(topicFilePath, pageIndex)
	if topics != nil && len(*topics) > 0 {
		c.Data["TopicsList"] = topics
		//c.Data["NoMore"] = false
	} else {
		c.Data["Info"] = "您没有发布过任何帖子"
	}

	c.TplName = "userinfo.tpl"
}

//编辑文章
type EditController struct {
	beego.Controller
}

func (c *EditController) Get() {
	c.Data["Navigation"] = navigation.GetNav()
	v := c.GetSession("user")
	if v == nil {
		c.TplName = "auth.tpl"
	} else {

		sid, _ := c.GetInt("sid", -1)
		aid, _ := c.GetInt("aid", -1)

		if sid == -1 || aid == -1 || !subject.IsSubjectIdExist(sid) {
			logs.Error("EditController err", sid, aid)
			c.Abort("401")
			return
		}

		ret, u := sqlsys.ReadSubjectFromDb(sid, aid)
		if 0 == ret {
			c.Data["TopicType"] = subject.GetTopicTyleList()
			c.Data["UserId"] = u.UserId
			c.Data["UserName"] = u.UserName
			c.Data["Title"] = u.Title
			c.Data["Sid"] = sid
			c.Data["ArtiId"] = aid

			if u.Anonymity {
				c.Data["Check"] = "checked"
			}

			if u.Path == "" {
				c.Data["Story"] = "很遗憾，用户没有留下TA的故事"
			} else {
				path := fmt.Sprintf("%s/%s", audit.ArticleDir, u.Path)
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
			c.TplName = "edit.tpl"

		}
	}
}

//管理中心
type ManageController struct {
	beego.Controller
}

func (c *ManageController) Get() {

	v := c.GetSession("id")
	if v == nil || !audit.IsAdmin(v.(int)) {
		//logs.Error("Id 是", v.(uint32))
		c.Abort("401")
	}

	u := c.GetSession("user")
	if u == nil {
		c.Abort("401")
	}

	c.Data["Navigation"] = navigation.GetNav()
	c.Data["ManageUrl"] = audit.GetCommonStrCfg("managerurl")
	c.Data["SubType"] = subject.GetMainPageSubjectData()

	sid, _ := c.GetInt("sid", -1)
	if sid == -1 {
		//不是查询结果，走到欢迎页面
		c.TplName = "manage.tpl"
		return
	}

	pageIndex, _ := c.GetInt("page", 0)

	name := c.GetString("name", "")
	if name != "" {
		user := &sqlsys.User{Name: name}

		if !user.GetUserByName() {
			c.Data["Info"] = fmt.Sprintf("[1]用户%s没有发布过任何帖子", name)
			c.TplName = "manage.tpl"
			return
		}
		offset := pageIndex * global.OnePageElementCount

		topics, nums := (*sqlsys.Subject)(nil).GetTopicListByFiledWithOffset("user_name", name, sid, offset, global.OnePageElementCount)
		if topics != nil && len(*topics) > 0 {
			c.Data["TopicsList"] = topics
			c.Data["Sid"] = sid
			c.Data["Info"] = fmt.Sprintf("用户 [%s] 第 %d 页帖子查找成功", name, pageIndex)

			//设置导航条
			urlPrex := fmt.Sprintf("%s?sid=%d&name=%s", audit.GetCommonStrCfg("managerurl"), sid, name)
			records, prev, next := global.CreateNavIndexByNums(pageIndex, nums, urlPrex, "&page")
			if records != nil {
				c.Data["RecordIndexs"] = records
				c.Data["PrePage"] = prev
				c.Data["NextPage"] = next
			}
		} else {
			c.Data["Info"] = fmt.Sprintf("[1]用户%s没有更多的帖子", name)
		}
		c.TplName = "manage.tpl"
		return
	}

	date := c.GetString("date", "")
	if date != "" {
		offset := pageIndex * global.OnePageElementCount
		topics, nums := (*sqlsys.Subject)(nil).GetTopicListByFiledWithOffset("date", date, sid, offset, global.OnePageElementCount)
		if topics != nil && len(*topics) > 0 {
			c.Data["TopicsList"] = topics
			c.Data["Sid"] = sid
			c.Data["Info"] = fmt.Sprintf("日期 [%s] 第 %d 页帖子查找成功", date, pageIndex)

			//设置导航条
			urlPrex := fmt.Sprintf("%s?sid=%d&date=%s", audit.GetCommonStrCfg("managerurl"), sid, date)
			records, prev, next := global.CreateNavIndexByNums(pageIndex, nums, urlPrex, "&page")
			if records != nil {
				c.Data["RecordIndexs"] = records
				c.Data["PrePage"] = prev
				c.Data["NextPage"] = next
			}

		} else {
			c.Data["Info"] = fmt.Sprintf("[1]日期%s没有更多的帖子", date)
		}
		c.TplName = "manage.tpl"
	}

}

type ManageData struct {
	Subid int    `form:"subid_" valid:"Required“`
	Type  int    `form:"type_" valid:"Required“`
	Key   string `form:"key_" valid:"Required“`
}

func (c *ManageController) Post() {
	v := c.GetSession("id")
	if v == nil || !audit.IsAdmin(v.(int)) {
		c.Ctx.WriteString("[1]没有权限")
		return
	}

	u := ManageData{}
	if err := c.ParseForm(&u); err != nil {
		c.Ctx.WriteString("[1]格式不对，请修正！")
	} else {
		//beego.Informational(u)
		if !DealParameterCheck(u, "[1]数据格式异常，请修正！", &c.Controller) {
			return
		}

		if !subject.IsSubjectIdExist(u.Subid) {
			c.Ctx.WriteString("[1]版块格式不对，请修正！")
			return
		}

		c.Data["Navigation"] = navigation.GetNav()
		c.Data["SubType"] = subject.GetMainPageSubjectData()

		if u.Type == 1 {
			ret := fmt.Sprintf("[0]/%s?sid=%d&name=%s", audit.GetCommonStrCfg("managerurl"), u.Subid, u.Key)
			c.Ctx.WriteString(ret)

		} else if u.Type == 2 {
			ret := fmt.Sprintf("[0]/%s?sid=%d&date=%s", audit.GetCommonStrCfg("managerurl"), u.Subid, u.Key)
			c.Ctx.WriteString(ret)
		}
	}
}

//禁用帖子
type DisableController struct {
	beego.Controller
}

type DisableData struct {
	Subid int `form:"subid_" valid:"Required“`
	Aid   int `form:"aid_" valid:"Required“`
}

func (c *DisableController) Post() {

	v := c.GetSession("id")
	if v == nil || !audit.IsAdmin(v.(int)) {
		c.Ctx.WriteString("[1]没有权限")
		return
	}

	u := DisableData{}
	if err := c.ParseForm(&u); err != nil {
		c.Ctx.WriteString("[1]格式不对，请修正！")
	} else {
		if !DealParameterCheck(u, "[1]数据格式异常，请修正！", &c.Controller) {
			return
		}

		if !subject.IsSubjectIdExist(u.Subid) {
			logs.Error("DisableController err", u.Subid)
			c.Ctx.WriteString("[1]找不到对应的帖子")
			return
		}

		aSubject := &sqlsys.Subject{Id: u.Aid}

		ok, status := aSubject.UpdateDisableStatus(u.Subid)
		if ok {
			if status {
				c.Ctx.WriteString("[0]文章禁用成功！")
				//fmt.Printf("1 %d %d\n", u.Aid, u.Subid)
			} else {
				c.Ctx.WriteString("[0]文章取消禁用成功！")
				//fmt.Printf("2 %d %d\n", u.Aid, u.Subid)
			}
			//clearcache.ClearPath(fmt.Sprintf("/cure%d", numId))
			//fmt.Printf("3 %d %d\n", u.Aid, u.Subid)
		} else {
			c.Ctx.WriteString("[1]文章设置禁用状态失败！")
			//fmt.Printf("4 %d %d\n", u.Aid, u.Subid)
		}

		//fmt.Printf("%d %d\n", u.Aid, u.Subid)
	}
}

//用户中心
type UserController struct {
	beego.Controller
}

type userMsg struct {
	UserId int `form:"userId_" valid:"Required“`
	Type   int `form:"type_" valid:"Required“`
}

func (c *UserController) Post() {
	v := c.GetSession("id")
	if v == nil || !audit.IsAdmin(v.(int)) {
		c.Ctx.WriteString("[1]没有权限")
		return
	}

	u := userMsg{}
	if err := c.ParseForm(&u); err != nil {
		c.Ctx.WriteString("[1]格式不对，请修正！")
	} else {
		if !DealParameterCheck(u, "[1]数据格式异常，请修正！", &c.Controller) {
			return
		}

		aUserAu := &sqlsys.UserAudit{UserId: u.UserId}

		if aUserAu.ReadDb() {
			if u.Type == 1 {
				//禁言
				aUserAu.Could = !aUserAu.Could
				aUserAu.UpdataCould()
				if aUserAu.Could {
					c.Ctx.WriteString("[0]禁言用户成功")
				} else {
					c.Ctx.WriteString("[0]取消禁言用户成功")
				}

			} else if u.Type == 2 {
				//提升等级
				aUserAu.Level++
				aUserAu.UpdataLevel()
				c.Ctx.WriteString("[0]提升用户等级成功")
			}
		} else {
			c.Ctx.WriteString("[1]操作失败")
		}
	}
}

func (c *UserController) Get() {
	c.Data["Navigation"] = navigation.GetNav()
	v := c.GetSession("id")
	if v == nil {
		c.TplName = "auth.tpl"
		return
	}

	name := c.GetString("name", "")
	if name == "" {
		c.Abort("401")
	}

	aUser := &sqlsys.User{Name: name}

	if aUser.GetUserByName() {
		aUserAu := &sqlsys.UserAudit{UserId: aUser.Id}
		c.Data["Name"] = aUser.Name
		c.Data["Birth"] = aUser.Birth
		c.Data["Lastlog"] = aUser.Lastlog

		if aUserAu.ReadDb() {
			c.Data["TlArticleNums"] = aUserAu.TlArticleNums
			c.Data["TlCommentTimes"] = aUserAu.TlCommentTimes
			c.Data["Level"] = aUserAu.Level
			if aUserAu.Could {
				c.Data["Status"] = "被禁言"
			} else {
				c.Data["Status"] = "正常"
			}
		}

		if audit.IsAdmin(v.(int)) {
			if aUser.Mail == "" {
				c.Data["Mail"] = "用户没有留下邮箱"
			} else {
				c.Data["Mail"] = aUser.Mail
			}

			c.Data["IsAdmin"] = true
			c.Data["UserId"] = aUserAu.UserId
		} else {
			c.Data["Mail"] = "仅限管理员可见"
		}
	}

	c.TplName = "user.tpl"
}
