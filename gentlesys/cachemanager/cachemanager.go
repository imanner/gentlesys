package cachemanager

import (
	"container/list"
	"fmt"
	"gentlesys/global"
	"gentlesys/models/nginx"
	"gentlesys/models/sqlsys"
	"gentlesys/subject"
	"gentlesys/timework"
	"net/http"
	"sort"
	"sync"
)

//一个用来管理缓存的文件，主要是对主题中的页面进行缓存，避免过多频繁查询数据库

func init() {
	//注意顺序
	initCacheSubject()

	//定时更新热点帖子。
	timework.AddPeriodicMinTask("hot", func() {
		for _, v := range CacheSubjectObjMaps {
			v.initCacheHotTopicListFromDb()
			//这里还有一个清除缓存
		}
	})
	//注册定时更新Nginx阅读量
	if global.IsNginxCache {
		timework.AddPeriodicMinTask("nginx", func() {
			nginx.UpdateTopicDayAccessTimes(UpdateCacheSubjectReadTimesWithNginx)
			nginx.ClearccessTimes()
			//刷新nginx对应的缓存页面
			clearNgnixDirtySubjectPages()
		})
	}
}

var dirtySubjectPages map[int]bool

//与板块话题相关的缓存区
var CacheSubjectObjMaps map[int]*CacheObj

//每个主题的global.CachePagesNums个页面都是在内存中的，每页global.OnePageElementCount个帖子
func initCacheSubject() {
	//subNodes := subject.GetMainPageSubjectData()
	subNodesMap := subject.GetSubjectMap()

	if global.IsNginxCache {
		dirtySubjectPages = make(map[int]bool, len(*subNodesMap))
	}

	CacheSubjectObjMaps = make(map[int]*CacheObj)
	for _, k := range *subNodesMap {
		CacheSubjectObjMaps[k.UniqueId] = new(CacheObj)
		CacheSubjectObjMaps[k.UniqueId].SubId = k.UniqueId
		CacheSubjectObjMaps[k.UniqueId].elementMap = make(map[int]*subjectNode)
		//CacheSubjectObjMaps[k.UniqueId].hotEleMap = make(map[int]*sqlsys.Subject)

		if k.UniqueId != 1001 {
			//最多10条最新通知，而通知主题本身不需要
			//CacheSubjectObjMaps[k.UniqueId].notices = make([]*sqlsys.Subject, 10)
			CacheSubjectObjMaps[k.UniqueId].notices = list.New()
		}
		CacheSubjectObjMaps[k.UniqueId].accessFlag = make([]int, global.OnePageElementCount*global.CachePagesNums)
		CacheSubjectObjMaps[k.UniqueId].initCacheTopicListFromDb()

		//最开始所有的页面都是干净的
		if global.IsNginxCache {
			dirtySubjectPages[k.UniqueId] = false
		}

	}
	//再初始化notices
	for _, k := range *subNodesMap {
		CacheSubjectObjMaps[k.UniqueId].initCacheNoticesList()
	}

}

type subjectNode struct {
	s     *sqlsys.Subject //实际元素
	times int             //修改过的次数
}
type CacheObj struct {
	SubId        int                  //所属的主题板块
	mutex        sync.RWMutex         //用于保护结构体的锁，保护整个结构
	newCount     int                  //新加元素的计数，为了配合nginx的缓存机制
	elementMap   map[int]*subjectNode //不使用[]，使用map，因为需要使用aid来快速定位到subject
	accessFlag   []int                //页面是否访问过的标识，如果是0，表示没有访问过
	notices      *list.List           //通知栏
	mutexNotices sync.RWMutex         //仅仅保护通知栏

	hotMutex    sync.RWMutex      //单独保护hotEleSlice的锁
	hotEleSlice []*sqlsys.Subject //热帖记录，仅仅缓存热帖
}

//更新缓存中的禁用状态。不更新数据库
func (c *CacheObj) UpdateSubjectDisableStatus(v *sqlsys.Subject) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.elementMap[v.Id]; ok {
		c.elementMap[v.Id].s.Disable = !c.elementMap[v.Id].s.Disable
	}
}

//获取主题的通知
func (c *CacheObj) GetNotices() *[]*sqlsys.Subject {
	if c.SubId == 1001 || c.notices.Len() == 0 {
		return nil
	}
	ret := make([]*sqlsys.Subject, c.notices.Len())
	i := 0

	c.mutexNotices.RLock()
	defer c.mutexNotices.RUnlock()

	for e := c.notices.Front(); e != nil; e = e.Next() {
		ret[i] = e.Value.(*sqlsys.Subject)
		i++
	}
	return &ret
}

//这个必须要在InitCacheTopicListFromDb 1001之后才能调用
func (c *CacheObj) initCacheNoticesList() {
	if c.SubId == 1001 {
		return
	}
	nums := subject.GetCurTotalTopicNums(1001)
	j := 0
	//notices的type表示其所在的主题id
	for i := nums; i >= 1; i-- {
		if v, ok := CacheSubjectObjMaps[1001].elementMap[i]; ok && v.s.Type == c.SubId {
			c.notices.PushBack(v.s)
			j++
			if j >= global.MaxNoticeShowNums {
				break
			}
		}
	}
}

//写一个按照阅读数量排序的的方法
type subjectSort struct {
	nodes []*sqlsys.Subject
}

func (s *subjectSort) Len() int {
	return len(s.nodes)
}
func (s *subjectSort) Swap(i, j int) {
	s.nodes[i], s.nodes[j] = s.nodes[j], s.nodes[i]
}
func (s *subjectSort) Less(i, j int) bool {
	return s.nodes[i].ReadTimes > s.nodes[j].ReadTimes
}

//初始化最高热度帖子的函数
func (c *CacheObj) initCacheHotTopicListFromDb() {
	//读取热帖
	tmpMap := make(map[int]bool)
	//先将elementMap中的20大热帖取出
	var topicList subjectSort
	//从elementMap读取帖子到切片
	c.mutex.RLock()
	topicList.nodes = make([]*sqlsys.Subject, len(c.elementMap))
	t := 0
	for _, v := range c.elementMap {
		topicList.nodes[t] = v.s
		t++
	}
	c.mutex.RUnlock()

	sort.Sort(&topicList)

	latelyLens := 20
	if latelyLens > len(topicList.nodes) {
		latelyLens = len(topicList.nodes)
	}

	nums := global.OnePageElementCount * global.CacheHotPagesNums
	//按照帖子热度获取帖子
	pHotTopicList := (*sqlsys.Subject)(nil).GetTopicListSortByField(c.SubId, "read_times", nums)

	hotLens := latelyLens + len(*pHotTopicList)

	//这里每次直接把hotEleSlice重新赋值
	c.hotMutex.Lock()
	defer c.hotMutex.Unlock()

	c.hotEleSlice = make([]*sqlsys.Subject, hotLens)

	//先获取最近的热点帖子
	j := 0
	for j = 0; j < latelyLens; j++ {
		c.hotEleSlice[j] = topicList.nodes[j]
		tmpMap[topicList.nodes[j].Id] = true
		//fmt.Printf("%d ", topicList.nodes[j].Id)
	}

	//再获取历史最高的热点帖子
	if len(*pHotTopicList) > latelyLens {
		for i, v := range *pHotTopicList {
			//去掉重复的
			if _, exist := tmpMap[v.Id]; !exist {
				c.hotEleSlice[j] = v
				j++
				//处理匿名
				if (*pHotTopicList)[i].Anonymity {
					(*pHotTopicList)[i].UserName = "匿名网友"
				}
			}
		}
	}
	c.hotEleSlice = c.hotEleSlice[0:j]
}

//初始化操作时没有加锁，考虑到还在程序初始化期，不会并发访问
func (c *CacheObj) initCacheTopicListFromDb() {
	nums := global.OnePageElementCount * global.CachePagesNums
	//按照发布时间获取帖子
	pTopicList := (*sqlsys.Subject)(nil).GetTopicListSortByField(c.SubId, "-id", nums)

	//将数据保存在列表中topicList 是 *[]orm.Params
	if len(*pTopicList) > 0 {
		for i, v := range *pTopicList {
			c.elementMap[v.Id] = &subjectNode{s: (*pTopicList)[i]}
			//处理匿名
			if (*pTopicList)[i].Anonymity {
				(*pTopicList)[i].UserName = "匿名网友"
			}
		}
		subject.UpdateCurTopicIndex(c.SubId, (*pTopicList)[0].Id)
	}
	//初始化热帖
	c.initCacheHotTopicListFromDb()
}

func UpdateCacheSubjectReadTimesWithNginx(sid int, aid int, times int) {
	CacheSubjectObjMaps[sid].updateCacheSubjectReadTimesWithNginx(aid, times)
}

//走到这里，肯定是配置了nginx缓存的
func (c *CacheObj) updateCacheSubjectReadTimesWithNginx(aid int, times int) {
	c.mutex.Lock()
	if _, ok := c.elementMap[aid]; ok {
		c.elementMap[aid].s.ReadTimes += times
		c.elementMap[aid].times += times

		//及时持久化nginx的访问量
		if c.elementMap[aid].times > 10 {
			c.elementMap[aid].s.UpdateSubjectField(c.SubId, "ReadTimes", "ReplyTimes")
			c.elementMap[aid].times = 0
		}

		c.mutex.Unlock()
	} else {
		c.mutex.Unlock()
		//否则更新数据库
		sqlsys.SubjectReadTimesUpdate(c.SubId, aid, times)
	}
	dirtySubjectPages[c.SubId] = true
}

func (c *CacheObj) UpdateCacheSubjectTimesField(v *sqlsys.Subject, field ...string) {
	//如果缓存中存在，则更新缓存的数据;后面在统一时机一次性更新数据库
	c.mutex.Lock()
	if _, ok := c.elementMap[v.Id]; ok {
		c.elementMap[v.Id].s.ReadTimes = v.ReadTimes
		c.elementMap[v.Id].s.ReplyTimes = v.ReplyTimes
		c.elementMap[v.Id].times++

		//只持久化变化次数超过10次以上的
		if c.elementMap[v.Id].times > 10 {
			c.elementMap[v.Id].s.UpdateSubjectField(c.SubId, "ReadTimes", "ReplyTimes")
			c.elementMap[v.Id].times = 0
		}

		c.mutex.Unlock()
		//fmt.Printf("更新缓存...\n")
	} else {
		c.mutex.Unlock()
		//否则更新数据库
		v.UpdateSubjectField(c.SubId, field...)
		//fmt.Printf("更新数据库...\n")
	}
}

//把缓存数据持久化到数据库，主要是读与回复的数量,这样避免每次都读写数据库，提高速度
func (c *CacheObj) saveCacheElement() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, v := range c.elementMap {
		//只持久化变化次数超过10次以上的
		if v.times > 10 {
			v.s.UpdateSubjectField(c.SubId, "ReadTimes", "ReplyTimes")
			v.times = 0
			//fmt.Printf("开始持久化 %d %d...\n", c.SubId, v.s.Id)
		}

	}
}

func (c *CacheObj) ReadSubjectFromCache(id int) (int, *sqlsys.Subject) {
	c.mutex.RLock()
	if v, ok := c.elementMap[id]; ok {
		//fmt.Printf("读取缓存...\n")
		c.mutex.RUnlock()
		return 0, v.s
	} else {
		//读取更新数据库
		//fmt.Printf("读取数据库...\n")
		c.mutex.RUnlock()
		return sqlsys.ReadSubjectFromDb(c.SubId, id)
	}

}

func (c *CacheObj) UpdateNoticeElement(v interface{}) {
	if c.SubId != 1001 {
		c.mutexNotices.Lock()
		defer c.mutexNotices.Unlock()
		c.notices.PushFront(v)
		if c.notices.Len() > global.MaxNoticeShowNums {
			e := c.notices.Back()
			c.notices.Remove(e)
		}
	}
}

//在链表头部加入一个数据。v 必须是 * 类型。返回值 是否清空过nginx缓存
func (c *CacheObj) AddElement(v interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if v.(*sqlsys.Subject).Anonymity {
		v.(*sqlsys.Subject).UserName = "匿名网友"
	}
	c.elementMap[v.(*sqlsys.Subject).Id] = &subjectNode{s: v.(*sqlsys.Subject)}
	c.newCount++

	if c.SubId == 1001 && v.(*sqlsys.Subject).Type != 1001 {
		//如果是公告，则还需要刷新对应主板公告的列表
		CacheSubjectObjMaps[v.(*sqlsys.Subject).Type].UpdateNoticeElement(v)
	}

	//刷新的条件：1 如果新发帖大于FlushNumsLimit，则删除elementMap尾部多余缓存的元素，避免map爆炸
	if c.newCount >= global.FlushNumsLimit {
		c.newCount = 0
		//删除最后的FlushNumsLimit个元素
		if len(c.elementMap) > global.CachePagesNums*global.OnePageElementCount {
			curTopicIndex := subject.GetCurTotalTopicNums(c.SubId)
			end := curTopicIndex - global.CachePagesNums*global.OnePageElementCount
			start := end - global.FlushNumsLimit
			for i := start; i <= end; i++ {
				//这里需要将元素持久化数据库。考虑用一个go程去做，避免长时间不返回
				if _, ok := c.elementMap[i]; ok {
					if c.elementMap[i].times > 0 {
						c.elementMap[i].s.UpdateSubjectField(c.SubId, "ReadTimes", "ReplyTimes")
					}
					delete(c.elementMap, i)
				}

			}
			//fmt.Printf("删除多余元素后，现在元素个数%d\n", len(c.elementMap))
		}
		//如果启用了nginx缓存，还需要将旧的页面清除
		if global.IsNginxCache {
			c.clearMainPageNginxCache()
		}

	}
}

func ClearNgnixCachePage(path string) {
	//fmt.Printf("clear ... %s\n", path)
	http.Head(fmt.Sprintf("http://127.0.0.1/pre%s", path))
}

func clearNgnixDirtySubjectPages() {
	for k, v := range dirtySubjectPages {
		if v {
			dirtySubjectPages[k] = false
			//这里只刷新了主页，而后面的页面都没有刷新，有一定影响页面阅读数的问题。
			ClearNgnixCachePage(fmt.Sprintf("/subject%d", k))
		}
	}
}

func ClearNgnixCachePageWithId(sid int, aid int, page int) {
	var url string
	if page == 0 {
		url = fmt.Sprintf("http://127.0.0.1/pre/browse?sid=%d&aid=%d", sid, aid)
	} else {
		url = fmt.Sprintf("http://127.0.0.1/pre/browse?sid=%d&aid=%d&page=%d", sid, aid, page)
	}
	http.Head(url)
}

func (c *CacheObj) clearMainPageNginxCache() {
	//不刷新page=0，因为没有page=0，page=0是主页
	ClearNgnixCachePage(fmt.Sprintf("/subject%d", c.SubId)) // 这个就是page=0
	for i := 1; i < global.CachePagesNums; i++ {
		//只有访问过的才清除
		if c.accessFlag[i] > 0 {
			ClearNgnixCachePage(fmt.Sprintf("/subject%d?page=%d", c.SubId, i))
			c.accessFlag[i] = 0
		}

	}
}

//读取一页的元素数据。一共有global.CachePagesNums页
func (c *CacheObj) ReadElementsWithPageNums(pageNums int) []*sqlsys.Subject {

	if pageNums < 0 || pageNums >= global.CachePagesNums {
		return nil
	}

	//注意要放在defer c.mutex.Unlock()的上面，因为refreshCacheElement也需要拿锁，否则死锁

	curTopicIndex := subject.GetCurTotalTopicNums(c.SubId)

	//先读出来c.newCount，避免中途二次读取c.newCount发生改变。尽量减小锁的粒度
	c.mutex.RLock()
	newCount := c.newCount
	c.mutex.RUnlock()

	end := curTopicIndex - newCount - pageNums*global.OnePageElementCount
	if end < 0 {
		return nil
	}
	start := end - global.OnePageElementCount //不包含start
	if start < 0 {
		start = 0
	}

	//如果访问的是首页，则还有加上addList。故首页可能不止50条记录。
	var ret []*sqlsys.Subject
	j := 0

	if pageNums == 0 {
		ret = make([]*sqlsys.Subject, end-start+newCount)
		end += newCount
	} else {
		//非首页不需要访问头部新加的元素
		ret = make([]*sqlsys.Subject, end-start)
	}

	//这里加mutex读锁，是因为合并中会访问，可能导致链表结构乱
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for index := end; index > start; index-- {
		if v, ok := c.elementMap[index]; ok {
			ret[j] = v.s
			j++
		}
	}

	//表示该页有被访问过,访问的次数超过设定值后，持久化缓存到数据库
	c.accessFlag[pageNums]++
	if c.accessFlag[pageNums] >= global.AccessTimesLimit {
		go c.saveCacheElement()
	}

	return ret
}

func (c *CacheObj) GetHotTopicCount() int {
	c.hotMutex.RLock()
	defer c.hotMutex.RUnlock()
	return len(c.hotEleSlice)
}

//读取热点帖子
func (c *CacheObj) ReadHotWithPageNums(pageNums int) ([]*sqlsys.Subject, int) {

	//多半页的最近热帖，所以一共最多有global.CacheHotPagesNums+1页
	if pageNums < 0 || pageNums > global.CacheHotPagesNums {
		return nil, 0
	}

	start := pageNums * global.OnePageElementCount

	c.hotMutex.RLock()
	defer c.hotMutex.RUnlock()

	if start > len(c.hotEleSlice) {
		return nil, 0
	}
	end := (pageNums + 1) * global.OnePageElementCount
	if end > len(c.hotEleSlice) {
		end = len(c.hotEleSlice)
	}
	//返回原始切片的一部分，原始切片里面存放的是地址，该地址值不会主动释放，故不会访问到空指针问题
	return c.hotEleSlice[start:end], len(c.hotEleSlice)
}
