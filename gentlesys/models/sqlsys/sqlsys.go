package sqlsys

import (
	"fmt"
	"gentlesys/global"
	"gentlesys/models/audit"
	"gentlesys/models/reg"
	"gentlesys/subject"
	"io/ioutil"
	"time"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

//用户的表
type User struct {
	Id      int       `orm:"unique"`                                                          //用户ID                                                    //ID
	Name    string    `orm:"size(32)" form:"name_" valid:"Required;MinSize(1);MaxSize(32)“`   //名称
	Passwd  string    `orm:"size(32)" form:"passwd_" valid:"Required;MinSize(6);MaxSize(32)“` //密码
	Birth   time.Time `orm:"size(12);auto_now_add;type(date)"`                                //注册时间
	Lastlog time.Time `orm:"size(12);auto_now;null;type(date)"`                               //上次登录时间
	Fail    int32     `orm:"null;"`                                                           //登录失败的次数                                           //连续登录失败的次数，做安全防护
	Mail    string    `form:"mail_"`                                                          //禁止操作的天数
}

//用户记录行为的表,防止灌水等，做安全使用
type UserAudit struct {
	UserId          int  `orm:"unique;pk"`           //用户ID
	Could           bool `orm:"null;default(false)"` //是否禁用该用户发言或点评
	TlCommentTimes  int  `orm:"null;"`               //总共评论的次数
	DayCommentTimes int  `orm:"null;"`               //今天评论的次数
	TlArticleNums   int  `orm:"null;"`               //总共发布文章的次数
	DayArticleNums  int  `orm:"null;"`               //今天发布文章的次数
}

func (v *UserAudit) IsAdmin() bool {
	return audit.IsAdmin(v.UserId)
}

func (v *UserAudit) UpdataDayArticle() bool {
	o := orm.NewOrm()
	if _, err := o.Update(v, "TlArticleNums", "DayArticleNums"); err == nil {
		return true
	}
	return false
}

//在审计中获取该用户的信息，有则返回成功
func (v *UserAudit) ReadDb() bool {
	o := orm.NewOrm()
	err := o.Read(v)

	if err == orm.ErrNoRows {
		//logs.Error(err, "查询不到")
		return false
	} else if err == orm.ErrMissPK {
		//logs.Error(err, "找不到主键")
		return false
	}
	return true
}

//插入一条记录
func (v *UserAudit) Insert() bool {
	o := orm.NewOrm()
	_, err := o.Insert(v)
	if err == nil {
		return true
	}
	return false
}

//主题的表
type Subject struct {
	Id         int    `orm:"unique"` //文章ID,主键
	UserId     int    //作者ID
	UserName   string `orm:"size(32);null"`
	Date       string `orm:"size(32);null"`
	Type       int    `orm:"null;default(0)"`     //类型： 吐槽 话题 求助 炫耀 失望
	Title      string `orm:"size(128)"`           //帖子名称
	ReadTimes  int    `orm:"null;default(0)"`     //阅读数
	ReplyTimes int    `orm:"null;default(0)"`     //回复数
	Disable    bool   `orm:"null;default(false)"` //禁用该帖子
	Anonymity  bool   `orm:"null;default(false)"` //匿名发表
	Path       string `orm:"size(64)"`            //文章路径，相对路径
}

//返回主题上指定页的帖子列表，注意是倒序
func (v *Subject) GetTopicListPageNum(subId int, pages int) *[]orm.Params {

	end := int(subject.GetSubjectById(subId).CurTopicIndex) - pages*global.OnePageElementCount

	if end <= 0 {
		return nil
	}
	start := end - global.OnePageElementCount

	if start <= 0 {
		start = 0
	}
	//logs.Error("%d-%d\n", start, end)
	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(fmt.Sprintf("sub%d", subId))
	var posts []orm.Params
	qs.Filter("id__gte", start).Filter("id__lte", end).OrderBy("-id").Values(&posts, "Id", "UserName", "Data", "Type", "Title", "ReadTimes", "ReplyTimes", "Anonymity")

	return &posts
}

//从subx主题表中倒序读取一定数量的帖子
func (s *Subject) GetTopicListSortByTime(subId int, nums int) *[]Subject { //单纯的按照发布时间先后排序

	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(fmt.Sprintf("sub%d", subId))
	var posts []orm.ParamsList
	qs.OrderBy("-id").Limit(nums).ValuesList(&posts, "Id", "UserName", "Date", "Type", "Title", "ReadTimes", "ReplyTimes", "Anonymity")
	var ret []Subject = make([]Subject, len(posts))
	for i, k := range posts {
		ret[i].Id = int(k[0].(int64))
		ret[i].UserName = k[1].(string)
		ret[i].Date = k[2].(string)
		ret[i].Type = int(k[3].(int64))
		ret[i].Title = k[4].(string)
		ret[i].ReadTimes = int(k[5].(int64))
		ret[i].ReplyTimes = int(k[6].(int64))
		ret[i].Anonymity = k[7].(bool)
	}
	return &ret
}

/*从主题数据表中根据主题id找到该主题,1表示失败，0表示成功*/
func ReadSubjectFromDb(subId int, topicId int) (int, *Subject) {
	o := orm.NewOrm()

	subInstance := GetInstanceById(subId)

	subobj := subInstance.GetSubject()
	subobj.Id = topicId

	err := o.Read(subInstance)

	if err == orm.ErrNoRows {
		//logs.Error(err, "查询不到")
		return 1, nil
	} else if err == orm.ErrMissPK {
		//logs.Error(err, "找不到主键")
		return 1, nil
	}

	return 0, subobj
}

func registerDB() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	auth := global.GetStringFromCfg("mysql::auth", "")
	if auth != "" {
		orm.RegisterDataBase("default", "mysql", auth, 50)
	} else {
		panic("没有配置mysql的认证项...")
	}
	orm.RegisterModel(new(User), new(UserAudit))
	subs := subject.GetMainPageSubjectData()
	for _, v := range *subs {
		orm.RegisterModel(GetInstanceById(v.UniqueId))
	}

	//最后才能运行这个启动
	orm.RunSyncdb("default", false, true)

}

func init() {
	registerDB()
}

//发送文章，从客户端提交过来的数据
type CommitArticle struct {
	ArtiId    int    `form:"atId_"`                    //文章Id,如果是编辑则有，是新建则无
	SubId     int    `form:"subId_" valid:"Required“`  //主题id
	UserId    int    `form:"userId_" valid:"Required“` //用户id
	Type      int    `form:"type_"`                    //话题类型
	Anonymity bool   `form:"anonymity_"`               //是否匿名
	UserName  string `form:"userName_" valid:"MinSize(1);MaxSize(32)" `
	Title     string `form:"title_" valid:"MinSize(1);MaxSize(128)"`
	Story     string `form:"story_" valid:"MaxSize(1000000)"`
}

func (v *CommitArticle) WriteDb() (int, *Subject) {
	o := orm.NewOrm()
	aTopicInter := GetInstanceById(v.SubId)

	aTopic := aTopicInter.GetSubject()

	aTopic.UserId = v.UserId
	aTopic.UserName = v.UserName
	aTopic.Type = v.Type
	aTopic.Title = v.Title
	aTopic.Date = time.Now().Format("2006-01-02 15:04:05")
	aTopic.Anonymity = v.Anonymity

	id, err := o.Insert(aTopicInter)
	if err != nil {
		logs.Error(err, id)
		return 0, nil
	}

	aTopic.Path = fmt.Sprintf("s%d_a%d", v.SubId, aTopic.Id)

	//把文字写到磁盘，数据库只保存文章的路径
	path := fmt.Sprintf("%s/%s", audit.ArticleDir, aTopic.Path)

	//去掉kindeditor非法的字符
	v.Story = reg.DelErrorString(v.Story)

	//图片加上自动适配
	v.Story = reg.AddImagAutoClass(v.Story)

	err2 := ioutil.WriteFile(path, []byte(v.Story), 0644)
	if err2 != nil {
		logs.Error(err2, aTopic.Id)
	}

	//aTopic.Href = fmt.Sprintf("/browse?sid=%d&aid=%d", v.SubId, aTopic.Id)

	if _, err := o.Update(aTopicInter, "Path"); err != nil {
		return 0, nil
	}

	return aTopic.Id, aTopic
}
