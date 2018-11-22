package cachemanager

import (
	"container/list"
	"fmt"
	"gentlesys/global"
	"gentlesys/models/sqlsys"
	"gentlesys/subject"
	"sync"
)

//一个用来管理缓存的文件，主要是对主题中的页面进行缓存，避免过多频繁查询数据库

const maxNoticesCount = 10

func init() {
	//注意顺序
	initCacheSubject()
}

//与板块话题相关的缓存区
var CacheSubjectObjMaps map[int]*CacheObj

//每个主题的global.CachePagesNums个页面都是在内存中的，每页global.OnePageElementCount个帖子
func initCacheSubject() {
	//subNodes := subject.GetMainPageSubjectData()
	subNodesMap := subject.GetSubjectMap()

	CacheSubjectObjMaps = make(map[int]*CacheObj)
	for _, k := range *subNodesMap {
		CacheSubjectObjMaps[k.UniqueId] = new(CacheObj)
		CacheSubjectObjMaps[k.UniqueId].SubId = k.UniqueId
		CacheSubjectObjMaps[k.UniqueId].elementMap = make(map[int]*subjectNode)
		if k.UniqueId != 1001 {
			//最多10条最新通知，而通知主题本身不需要
			//CacheSubjectObjMaps[k.UniqueId].notices = make([]*sqlsys.Subject, 10)
			CacheSubjectObjMaps[k.UniqueId].notices = list.New()
		}
		CacheSubjectObjMaps[k.UniqueId].accessFlag = make([]int, global.OnePageElementCount*global.CachePagesNums)
		CacheSubjectObjMaps[k.UniqueId].InitCacheTopicListFromDb()
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
	SubId      int                  //所属的主题板块
	mutex      sync.RWMutex         //用于保护结构体的锁，保护list
	newCount   int                  //新加元素的计数，为了配合nginx的缓存机制
	elementMap map[int]*subjectNode //不使用[]，使用map，因为需要使用aid来快速定位到subject
	accessFlag []int                //页面是否访问过的标识，如果是0，表示没有访问过
	//notices    []*sqlsys.Subject    //通知栏
	notices      *list.List   //通知栏
	mutexNotices sync.RWMutex //仅仅保护通知栏
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
	for i := 1; i <= nums; i++ {
		if v, ok := CacheSubjectObjMaps[1001].elementMap[i]; ok && v.s.Type == c.SubId {
			//c.notices[j] = v.s
			c.notices.PushFront(v.s)
			j++
			if j >= maxNoticesCount {
				break
			}
		}
	}
}

//初始化操作时没有加锁，考虑到还在程序初始化期，不会并发访问
func (c *CacheObj) InitCacheTopicListFromDb() {
	//var aSubject sqlsys.Subject
	nums := global.OnePageElementCount * global.CachePagesNums
	pTopicList := (*sqlsys.Subject)(nil).GetTopicListSortByTime(c.SubId, nums)

	//将数据保存在列表中topicList 是 *[]orm.Params
	if len(*pTopicList) > 0 {
		for i, v := range *pTopicList {
			c.elementMap[v.Id] = &subjectNode{s: &(*pTopicList)[i]}
			//处理匿名
			if (*pTopicList)[i].Anonymity {
				(*pTopicList)[i].UserName = "匿名网友"
			}
		}

		subject.UpdateCurTopicIndex(c.SubId, (*pTopicList)[0].Id)
	}

	//atomic.StoreUint32(&mysqlTool.ShareCureIndex, shareList[0].Id)
}

func (c *CacheObj) UpdateCacheSubjectTimesField(v *sqlsys.Subject, field ...string) {
	//如果缓存中存在，则更新缓存的数据;后面在统一时机一次性更新数据库
	c.mutex.Lock()
	if _, ok := c.elementMap[v.Id]; ok {
		c.elementMap[v.Id].s.ReadTimes = v.ReadTimes
		c.elementMap[v.Id].s.ReplyTimes = v.ReplyTimes
		c.elementMap[v.Id].times++
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
		if c.notices.Len() > maxNoticesCount {
			e := c.notices.Back()
			c.notices.Remove(e)
		}
	}
}

//在链表头部加入一个数据。v 必须是 * 类型。返回值 是否清空过nginx缓存
func (c *CacheObj) AddElement(v interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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
			fmt.Printf("删除多余元素后，现在元素个数%d\n", len(c.elementMap))
		}

	}
}

/*
func (c *CacheObj) clearMainPageNginxCache() {
	//不刷新page=0，因为没有page=0，page=0是主页
	clearcache.ClearPath("/") // 这个就是page=0
	for i := 1; i < global.CachePagesNums; i++ {
		//只有访问过的才清除
		if c.accessFlag[i] == 1 {
			clearcache.ClearPath(fmt.Sprintf("/?page=%d", i))
			c.accessFlag[i] = 0
			//fmt.Printf("clear cache page %d ...\n", i)
		}

	}
	//fmt.Printf("clear cache ...\n")
}
*/
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
