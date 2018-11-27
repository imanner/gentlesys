package sqlsys

import (
	"crypto/md5"
	"fmt"
	"gentlesys/global"
	"gentlesys/models/audit"
	"gentlesys/models/reg"
	"gentlesys/subject"
	"gentlesys/timework"
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
	Fail    int       `orm:"null;"`
	//orm不能随便使用 default(""),不然正常值从mail_过来的，反而无法写入                                //登录失败的次数                                           //连续登录失败的次数，做安全防护
	Mail string `orm:"size(64)" form:"mail_" valid:"MaxSize(64)"` //邮箱
}

const ERR_NO_USERNAME = 1         //没有该用户
const ERR_AUTH_FAIL = 2           //认证失败
const ERR_USERNAME_NOT_UNIQUE = 3 //用户名不是唯一
const ERR_FAIL_FORBID = 4         //登录失败过多被锁定

//定期每日清理一些锁定字段
func sqlDailyClearStatus() {
	o := orm.NewOrm()
	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable("user")

	num, _ := qs.Filter("fail__gt", 0).Update(orm.Params{
		"fail": 0,
	})

	s := orm.NewOrm()
	// 获取 QuerySeter 对象，user 为表名
	qs1 := s.QueryTable("user_audit")

	num1, _ := qs1.Filter("day_comment_times__gt", 0).Update(orm.Params{
		"day_comment_times": 0,
	})

	num2, _ := qs1.Filter("day_article_nums__gt", 0).Update(orm.Params{
		"day_article_nums": 0,
	})

	logs.Error(fmt.Sprintf("今天共清理%d条记录", num+num1+num2))
}

func (v *User) GetUserByName() bool {
	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(v)

	err := qs.Filter("Name", v.Name).One(v)
	if err == orm.ErrMultiRows {
		return false
	}
	if err == orm.ErrNoRows {
		return false
	}

	//v.Id = oldUser.Id
	return true
}

func (v *User) UpdatePasswdByName() int {
	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(v)
	oldUser := &User{}
	err := qs.Filter("Name", v.Name).One(oldUser)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		logs.Error("更新密码：用户名不是唯一")
		return ERR_USERNAME_NOT_UNIQUE
	}
	if err == orm.ErrNoRows {
		// 没有找到记录
		logs.Error("更新密码：没有该用户", v.Name)
		return ERR_NO_USERNAME
	}

	v.Id = oldUser.Id

	if _, err := o.Update(v, "Passwd"); err == nil {
		return 0
	} else {
		logs.Error(err, "更新密码错误")
	}

	return 1
}

//通过名字寻找邮箱
func (v *User) FindMailByName() int {
	o := orm.NewOrm()
	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(v)
	err := qs.Filter("Name", v.Name).One(v)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		logs.Error(err, "用户名不是唯一")
		return ERR_USERNAME_NOT_UNIQUE
	}
	if err == orm.ErrNoRows {
		// 没有找到记录
		logs.Error(err, "没有该用户", v.Name)
		return ERR_NO_USERNAME
	}
	return 0
}

//检查用户名是否被使用
func (v *User) CheckUserExist() bool {
	o := orm.NewOrm()
	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable((*User)(nil))

	return qs.Filter("Name", v.Name).Exist()
}

//成功返回0
func (v *User) CheckUserAuth() int {

	var auser User

	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable("user")

	err := qs.Filter("Name", v.Name).One(&auser)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		logs.Error(err, "认证失败：用户名不是唯一")
		return ERR_USERNAME_NOT_UNIQUE
	}
	if err == orm.ErrNoRows {
		// 没有找到记录
		logs.Error("认证失败：没有该用户", v.Name)
		return ERR_NO_USERNAME
	}

	v.Id = auser.Id
	//错误次数过多，禁止登陆
	if auser.Fail > audit.GetCommonIntCfg("dayLogFailTime") {
		return ERR_FAIL_FORBID
	}

	if auser.Passwd == v.Passwd {
		//刷新一下登录时间
		o.Update(&auser, "lastlog")
		return 0
	}
	//存在用户，但是密码错误
	v.Fail = auser.Fail
	return ERR_AUTH_FAIL
}

func (v *User) WriteDb() int {
	o := orm.NewOrm()

	id, err := o.Insert(v)
	if err != nil {
		logs.Error(err, id)
		return 0
	}

	return int(id)
}

func (v *User) ReadDb() bool {
	o := orm.NewOrm()
	//aShare := Share{Id: id}

	err := o.Read(v)

	if err == orm.ErrNoRows {
		logs.Error(err, "查询不到")
		return false
	} else if err == orm.ErrMissPK {
		logs.Error(err, "找不到主键")
		return false
	}
	return true
}

//更新用户信息失败次数
func (v *User) UpdateFail() bool {
	o := orm.NewOrm()
	if _, err := o.Update(v, "Fail"); err == nil {
		return true
	} else {
		logs.Error(err, "更新错误")
	}

	return false
}

//用户记录行为的表,防止灌水等，做安全使用
type UserAudit struct {
	UserId          int  `orm:"unique;pk"`           //用户ID
	Could           bool `orm:"null;default(false)"` //是否禁用该用户发言或点评
	TlCommentTimes  int  `orm:"null;"`               //总共评论的次数
	DayCommentTimes int  `orm:"null;"`               //今天评论的次数
	TlArticleNums   int  `orm:"null;"`               //总共发布文章的次数
	DayArticleNums  int  `orm:"null;"`               //今天发布文章的次数
	Level           int  `orm:"null;default(1)"`     //级别或职位
}

func (v *UserAudit) IsAdmin() bool {
	return audit.IsAdmin(v.UserId)
}

func (v *UserAudit) UpdataCould() bool {
	o := orm.NewOrm()
	if _, err := o.Update(v, "Could"); err == nil {
		return true
	}
	return false
}

func (v *UserAudit) UpdataLevel() bool {
	o := orm.NewOrm()
	if _, err := o.Update(v, "Level"); err == nil {
		return true
	}
	return false
}

func (v *UserAudit) UpdataDayArticle() bool {
	o := orm.NewOrm()
	if _, err := o.Update(v, "TlArticleNums", "DayArticleNums"); err == nil {
		return true
	}
	return false
}

func (v *UserAudit) UpdataDayCommentTimes() bool {
	o := orm.NewOrm()
	if _, err := o.Update(v, "TlCommentTimes", "DayCommentTimes"); err == nil {
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
	Type       int    `orm:"null;default(0)"`     //类型： 吐槽 话题 求助 炫耀 失望//公告是表示subId
	Title      string `orm:"size(128)"`           //帖子名称
	ReadTimes  int    `orm:"null;default(0)"`     //阅读数
	ReplyTimes int    `orm:"null;default(0)"`     //回复数
	Disable    bool   `orm:"null;default(false)"` //禁用该帖子
	Anonymity  bool   `orm:"null;default(false)"` //匿名发表
	Path       string `orm:"size(64)"`            //文章路径，相对路径
}

func (s *Subject) GetArtiPath(subId int) string {
	return fmt.Sprintf("%s/s%d_a%d", audit.ArticleDir, subId, s.Id)
}

//更新帖子的阅读和评论数量
func UpdateTopicReadStatics(sid int, aid int, readTimes int, replyTimes int) bool {
	o := orm.NewOrm()

	//先根据subid artid读取记录
	subInstance := GetInstanceById(sid)

	subobj := subInstance.GetSubject()
	subobj.Id = aid
	subobj.ReadTimes = readTimes
	subobj.ReplyTimes = replyTimes

	if _, err := o.Update(subInstance, "ReadTimes", "ReplyTimes"); err != nil {
		return false
	}
	return true
}

//更新帖子的字段状态,这个就直接更新了，写入到数据库去
func (v *Subject) UpdateSubjectField(sid int, field ...string) bool {
	o := orm.NewOrm()

	//先根据subid artid读取记录
	subInstance := GetInstanceById(sid)

	subobj := subInstance.GetSubject()
	*subobj = *v

	if _, err := o.Update(subInstance, field...); err != nil {
		return false
	}
	return true
}

//更新帖子的禁用状态
func (v *Subject) UpdateDisableStatus(sid int) (bool, bool) {

	o := orm.NewOrm()

	//先根据subid artid读取记录
	subInstance := GetInstanceById(sid)

	subobj := subInstance.GetSubject()
	subobj.Id = v.Id

	//严重注意：这里有一个orm故障，如果写成o.Read(subInstance，"Disable")只去掉一项，经常会读取不到
	//而且就算读取到后，里面的subobj.Id会变成加一之后的值，非常诡异。怀疑是orm的bug
	err := o.Read(subInstance)
	if err == orm.ErrNoRows {
		return false, false
	} else if err == orm.ErrMissPK {
		return false, false
	}

	subobj.Disable = !subobj.Disable

	if _, err := o.Update(subInstance, "Disable"); err != nil {
		return false, subobj.Disable
	}
	return true, subobj.Disable
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
	qs.Filter("id__gte", start).Filter("id__lte", end).OrderBy("-id").Values(&posts, "Id", "UserName", "Date", "Type", "Title", "ReadTimes", "ReplyTimes", "Anonymity")

	return &posts
}

//从subx主题表读取一定数量的帖子，按照热度即阅读数排名
func (s *Subject) GetTopicListSortByField(subId int, field string, nums int) *[]*Subject { //单纯的按照发布时间先后排序

	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(fmt.Sprintf("sub%d", subId))
	var posts []orm.ParamsList
	qs.OrderBy(field).Limit(nums).ValuesList(&posts, "Id", "UserName", "Date", "Type", "Title", "ReadTimes", "ReplyTimes", "Disable", "Anonymity")
	var ret []*Subject = make([]*Subject, len(posts))
	for i, k := range posts {
		ret[i] = &Subject{}
		ret[i].Id = int(k[0].(int64))
		ret[i].UserName = k[1].(string)
		ret[i].Date = k[2].(string)
		ret[i].Type = int(k[3].(int64))
		ret[i].Title = k[4].(string)
		ret[i].ReadTimes = int(k[5].(int64))
		ret[i].ReplyTimes = int(k[6].(int64))
		ret[i].Disable = k[7].(bool)
		ret[i].Anonymity = k[8].(bool)
		ret[i].Path = fmt.Sprintf("s%d_a%d", subId, ret[i].Id)
	}
	return &ret
}

//从subx主题表中倒序读取一定数量的帖子
/*
func (s *Subject) GetTopicListSortByTime(subId int, nums int) *[]Subject { //单纯的按照发布时间先后排序

	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(fmt.Sprintf("sub%d", subId))
	var posts []orm.ParamsList
	qs.OrderBy("-id").Limit(nums).ValuesList(&posts, "Id", "UserName", "Date", "Type", "Title", "ReadTimes", "ReplyTimes", "Disable", "Anonymity")
	var ret []Subject = make([]Subject, len(posts))
	for i, k := range posts {
		ret[i].Id = int(k[0].(int64))
		ret[i].UserName = k[1].(string)
		ret[i].Date = k[2].(string)
		ret[i].Type = int(k[3].(int64))
		ret[i].Title = k[4].(string)
		ret[i].ReadTimes = int(k[5].(int64))
		ret[i].ReplyTimes = int(k[6].(int64))
		ret[i].Disable = k[7].(bool)
		ret[i].Anonymity = k[8].(bool)
		ret[i].Path = fmt.Sprintf("s%d_a%d", subId, ret[i].Id)
	}
	return &ret
}*/

//从subx主题表中根据字段名称查找帖子,从偏移offset开始
func (s *Subject) GetTopicListByFiledWithOffset(filed string, value string, subId int, offset int, limits int) (*[]Subject, int) { //单纯的按照发布时间先后排序

	o := orm.NewOrm()

	// 获取 QuerySeter 对象，user 为表名
	qs := o.QueryTable(fmt.Sprintf("sub%d", subId))
	var posts []orm.ParamsList

	key := fmt.Sprintf("%s__startswith", filed)

	cnt, _ := qs.Filter(key, value).Count()

	if offset > int(cnt) {
		return nil, 0
	}

	qs.Filter(key, value).OrderBy("-id").Offset(offset).Limit(limits).ValuesList(&posts, "Id", "UserName", "Date", "Title", "ReadTimes", "ReplyTimes", "Disable")
	var ret []Subject = make([]Subject, len(posts))
	for i, k := range posts {
		ret[i].Id = int(k[0].(int64))
		ret[i].UserName = k[1].(string)
		ret[i].Date = k[2].(string)
		ret[i].Title = k[3].(string)
		ret[i].ReadTimes = int(k[4].(int64))
		ret[i].ReplyTimes = int(k[5].(int64))
		ret[i].Disable = k[6].(bool)
	}
	return &ret, int(cnt)
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
	orm.RegisterModel(new(User), new(UserAudit), new(PasswdReset), new(Sub1001))
	subs := subject.GetMainPageSubjectData()
	for _, v := range *subs {
		orm.RegisterModel(GetInstanceById(v.UniqueId))
	}

	//最后才能运行这个启动
	orm.RunSyncdb("default", false, true)

}

func init() {
	registerDB()

	//注册定时清理任务
	timework.AddDailyTask("user", func() {
		sqlDailyClearStatus()
	})
}

//发送文章，从客户端提交过来的数据
type CommitArticle struct {
	ArtiId    int    `form:"atId_"`                    //文章Id,如果是编辑则有，是新建则无
	SubId     int    `form:"subId_" valid:"Required“`  //主题id
	UserId    int    `form:"userId_" valid:"Required“` //用户id
	Type      int    `form:"type_"`                    //话题类型
	Anonymity bool   `form:"anonymity_"`               //是否匿名
	UserName  string `form:"userName_" valid:"MinSize(1);MaxSize(32)" `
	Title     string `form:"title_" valid:"Required;MinSize(1);MaxSize(128)"`
	Story     string `form:"story_" valid:"Required;MaxSize(1000000)"`
}

func (v *CommitArticle) UpdateDb() bool {

	o := orm.NewOrm()

	//先根据subid artid读取记录
	subInstance := GetInstanceById(v.SubId)

	subobj := subInstance.GetSubject()
	subobj.Id = v.ArtiId
	//fmt.Printf("0 -- %d %d %d\n", v.ArtiId, subobj.Id, subInstance.GetSubject().Id)
	err := o.Read(subInstance)

	if err == orm.ErrNoRows {
		return false
	} else if err == orm.ErrMissPK {
		return false
	}

	//fmt.Printf("1 -- %d %d %d\n", v.ArtiId, subobj.Id, subInstance.GetSubject().Id)

	//对比二者的作者是不是同一个，否则不能篡改
	if v.UserId != subobj.UserId {
		logs.Error(fmt.Sprintf("帖子id %d 与用户id %d匹配不上", v.UserId, subobj.UserId))
		return false
	}

	subobj.Type = v.Type
	//subobj.Title = v.Title //题目不能修改，因为要同步修改缓存
	subobj.Anonymity = v.Anonymity

	if _, err := o.Update(subInstance, "Type", "Anonymity"); err != nil {
		return false
	}

	path := fmt.Sprintf("%s/%s", audit.ArticleDir, subobj.Path)

	//去掉kindeditor非法的字符
	v.Story = reg.DelErrorString(v.Story)

	//图片加上自动适配
	v.Story = reg.AddImagAutoClass(v.Story)

	err2 := ioutil.WriteFile(path, []byte(v.Story), 0644)
	if err2 != nil {
		logs.Error(err2)
		return false
	}

	return true
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

//用户重置密码的数据库相关字段
type PasswdReset struct {
	UserId string `orm:"unique;pk" valid:"MinSize(1);MaxSize(32)"` //用户名的md5
	Name   string `orm:"size(32)" valid:"Required;MinSize(1);MaxSize(32)"`
}

func (v *PasswdReset) Delete() {
	o := orm.NewOrm()
	o.Delete(v)
}

func (v *PasswdReset) ReadDb() bool {
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

//数据库插入值
func (v *PasswdReset) InsertByName() bool {

	t := time.Now()

	data := []byte(fmt.Sprintf("%s%d", v.Name, t.Unix()))
	mds := md5.Sum(data)
	o := orm.NewOrm()
	v.UserId = fmt.Sprintf("%x", mds) //将[]byte转成16进制
	id, err := o.Insert(v)
	if err != nil {
		logs.Error(err, id)
		return false
	}
	return true
}

//定时每天更新访问量
func SubjectReadTimesUpdate(sid int, aid int, times int) {
	//再依次更新帖子的访问量
	o := orm.NewOrm()
	table := fmt.Sprintf("sub%d", sid)
	o.QueryTable(table).Filter("id", aid).Update(orm.Params{
		"read_times": orm.ColValue(orm.ColAdd, times),
	})
	//fmt.Printf("sid %d aid %d add %d\n", sid, aid, times)
}
